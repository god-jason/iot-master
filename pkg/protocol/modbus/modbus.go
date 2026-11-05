// Package modbus Modbus 协议公共基础层
//
// TCP 和 RTU 共享：baseProtocol、pending 管理、寄存器/Mapper 解析辅助函数，
// 以及通用的 Read/Write/Poll 事务逻辑（通过 transport 接口屏蔽帧编码差异）。
package modbus

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

// reqCtx 一次请求-响应的上下文。
//
// 明确区分 TCP 的 UnitID 与 RTU 的从站地址 ——
// 此前二者共用 transport 的同一个 slave 参数，导致同名参数两种语义。
type reqCtx struct {
	txID   uint16 // TCP 事务 ID；RTU 无事务 ID，恒为 0
	unitID byte   // TCP UnitID
	slave  byte   // RTU 从站地址
}

// transport 抽象 TCP / RTU 在「帧编码 + pending 关联」上的差异。
// baseProtocol 的通用 Read/Write/Poll 只依赖本接口，从而消除两种协议间的重复实现。
type transport interface {
	// register 登记一个 pending 请求，返回请求上下文、响应通道和取消函数。
	// slave 由实现自行解释：TCP 下即 UnitID，RTU 下即从站地址。
	register(slave, fc byte) (reqCtx, chan *Frame, func())
	// encode 按协议编码请求帧；pdu 为纯 PDU（不含 FC），FC 由 fc 参数提供。
	encode(ctx reqCtx, fc byte, pdu []byte) []byte
}

// pendingRequest 一个待匹配的请求
type pendingRequest struct {
	ch chan *Frame
}

// baseProtocol TCP / RTU 共享状态
//
// Attach/Close/Mount/UnMount/Name/Poll/Read/Write 等通用行为都在本层实现，
// TCP/RTU 只需提供 transport（register/encode）与 OnData，不再各自重复透传方法。
type baseProtocol struct {
	name   string    // 协议名（protocol.TCP / protocol.RTU）
	tr     transport // 协议相关的帧编码 + pending 关联
	serial bool      // 是否串行化请求（RTU 半双工为 true）

	link     io.ReadWriteCloser
	onValues protocol.OnValues
	mu       sync.Mutex

	// 设备挂载（per-connection，因为 device 绑定在具体 TCP/RTU 连接上）
	devices sync.Map // deviceId → *ModbusDevice

	pendingMu sync.Mutex
	pending   map[string]*pendingRequest
	txID      uint16
	timeout   time.Duration

	serialMu sync.Mutex // serial=true 时串行化整个「发送 + 等待响应」过程
}

// newBaseProtocol 构造 baseProtocol。
// tr 通常是嵌入本结构的 TCPProtocol / RTUProtocol 自身（构造后回填）。
func newBaseProtocol(name string, tr transport, serial bool) baseProtocol {
	return baseProtocol{
		name:    name,
		tr:      tr,
		serial:  serial,
		pending: make(map[string]*pendingRequest),
		timeout: 3 * time.Second,
	}
}

func rtuPendingKey(slave, fc byte) string {
	return fmt.Sprintf("%02X:%02X", slave, fc)
}

// =====================================================================
// protocol.Protocol 通用实现（TCP / RTU 共用，子类型直接继承）
// =====================================================================

// Name 返回协议标识
func (b *baseProtocol) Name() string { return b.name }

// Attach 绑定到底层连接
func (b *baseProtocol) Attach(link io.ReadWriteCloser, onValues protocol.OnValues) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.link = link
	b.onValues = onValues
	return nil
}

// Close 释放所有资源
func (b *baseProtocol) Close() error {
	b.pendingMu.Lock()
	for _, p := range b.pending {
		close(p.ch)
	}
	b.pending = make(map[string]*pendingRequest)
	b.pendingMu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()
	b.devices = sync.Map{}
	b.link = nil
	return nil
}

// Mount 挂载子设备
func (b *baseProtocol) Mount(deviceId string, address string, modelId string) error {
	dev, err := NewDevice(deviceId, "", modelId, address)
	if err != nil {
		return err
	}
	b.devices.Store(deviceId, dev)
	return nil
}

// UnMount 卸载子设备
func (b *baseProtocol) UnMount(deviceId string) {
	b.devices.Delete(deviceId)
}

// Poll 轮询一个子设备所有点位
func (b *baseProtocol) Poll(deviceId string, timeout time.Duration) (map[string]any, error) {
	return b.poll(b.tr, deviceId, effectiveTimeout(b.timeout, timeout))
}

// Read 读取子设备某个点位
func (b *baseProtocol) Read(deviceId string, key string, timeout time.Duration) (any, error) {
	return b.read(b.tr, deviceId, key, effectiveTimeout(b.timeout, timeout))
}

// Write 写一个值到子设备某个点位
func (b *baseProtocol) Write(deviceId string, key string, value any, timeout time.Duration) error {
	return b.write(b.tr, deviceId, key, value, effectiveTimeout(b.timeout, timeout))
}

// =====================================================================
// 内部辅助
// =====================================================================

// writeBytes 写到底层连接（无锁，上层负责同步）
func (b *baseProtocol) writeBytes(data []byte) error {
	if b.link == nil {
		return fmt.Errorf("modbus: 底层连接未绑定")
	}
	_, err := b.link.Write(data)
	return err
}

// findDevice 按 deviceId 拿 *ModbusDevice
func (b *baseProtocol) findDevice(deviceId string) (*ModbusDevice, bool) {
	v, ok := b.devices.Load(deviceId)
	if !ok {
		return nil, false
	}
	return v.(*ModbusDevice), true
}

// findModel 按 modelId 从全局缓存拿 *ModbusModel
func (b *baseProtocol) findModel(modelId string) (*ModbusModel, bool) {
	v, ok := protocol.GetModel(modelId)
	if !ok {
		return nil, false
	}
	m, ok := v.(*ModbusModel)
	return m, ok
}

// registerPendingByKey 登记一个 pending 请求（通用，key 由 transport 决定）
func (b *baseProtocol) registerPendingByKey(key string) (chan *Frame, func()) {
	ch := make(chan *Frame, 1)
	b.pendingMu.Lock()
	b.pending[key] = &pendingRequest{ch: ch}
	b.pendingMu.Unlock()
	return ch, func() {
		b.pendingMu.Lock()
		delete(b.pending, key)
		b.pendingMu.Unlock()
	}
}

// =====================================================================
// 通用事务（Read / Write / Poll 共用）
// =====================================================================

// transact 发送一帧（pdu 不含 FC）并等待响应。
func (b *baseProtocol) transact(tr transport, slave, fc byte, pdu []byte, timeout time.Duration) (*Frame, error) {
	// RTU 半双工：串行化整个「登记 + 发送 + 等待响应」，避免 slave:fc 维度的 pending 相互覆盖。
	// 锁在本函数内成对获取/释放，不再跨 register / cancel 两个闭包（此前易漏调导致死锁）。
	if b.serial {
		b.serialMu.Lock()
		defer b.serialMu.Unlock()
	}

	ctx, ch, cancel := tr.register(slave, fc)
	defer cancel()

	data := tr.encode(ctx, fc, pdu)

	b.mu.Lock()
	err := b.writeBytes(data)
	b.mu.Unlock()
	if err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp == nil {
			return nil, protocol.ErrTimeout
		}
		return resp, nil
	case <-time.After(timeout):
		return nil, protocol.ErrTimeout
	}
}

// transactRead 发送读请求并回填响应上下文。
// 读响应帧本身不含 StartAddr/Quantity，需由请求上下文恢复，统一在此处完成。
func (b *baseProtocol) transactRead(tr transport, slave, fc byte, pdu []byte, startAddr, quantity uint16, timeout time.Duration) (*Frame, error) {
	resp, err := b.transact(tr, slave, fc, pdu, timeout)
	if err != nil {
		return nil, err
	}
	resp.StartAddr = startAddr
	resp.Quantity = quantity
	return resp, nil
}

// resolveModel 查找 device + model（Read / Write / Poll 共用）
func (b *baseProtocol) resolveModel(deviceId string) (*ModbusDevice, *ModbusModel, error) {
	dev, ok := b.findDevice(deviceId)
	if !ok {
		return nil, nil, protocol.ErrNotMounted
	}
	model, ok := b.findModel(dev.Model)
	if !ok {
		return nil, nil, fmt.Errorf("modbus: 设备 %s 的物模型 %s 未注册", deviceId, dev.Model)
	}
	return dev, model, nil
}

// read 读取单个 key（TCP / RTU 共用）
func (b *baseProtocol) read(tr transport, deviceId, key string, timeout time.Duration) (any, error) {
	dev, model, err := b.resolveModel(deviceId)
	if err != nil {
		return nil, err
	}
	target := model.FindItemByName(key)
	if target == nil {
		return nil, fmt.Errorf("modbus: key %q 未找到", key)
	}
	// 尽量复用包含 target 的批量读 poller，减少读请求；否则退化为单点读取
	fc, startAddr, quantity := target.FC, target.StartAddr, target.Quantity
	if p := model.findPollerFor(target); p != nil {
		fc, startAddr, quantity = p.FC(), p.StartAddr(), p.Quantity()
	}

	resp, err := b.transactRead(tr, dev.Slave, fc, BuildReadRequest(startAddr, quantity), startAddr, quantity, timeout)
	if err != nil {
		return nil, err
	}
	if resp.IsError {
		return nil, fmt.Errorf("modbus read error: %s", resp.ErrorMsg)
	}
	vals, err := model.DecodeResponse(dev.Slave, dev.Slave, resp)
	if err != nil {
		return nil, err
	}
	v, ok := vals[key]
	if !ok {
		return nil, fmt.Errorf("modbus: key %q 未在响应中找到", key)
	}
	return v, nil
}

// write 写入单个 key（TCP / RTU 共用）
func (b *baseProtocol) write(tr transport, deviceId, key string, value any, timeout time.Duration) error {
	dev, model, err := b.resolveModel(deviceId)
	if err != nil {
		return err
	}
	target := model.FindItemByName(key)
	if target == nil {
		return fmt.Errorf("modbus: key %q 未找到", key)
	}

	fc, pdu, err := buildWritePDU(target, value)
	if err != nil {
		return err
	}

	resp, err := b.transact(tr, dev.Slave, fc, pdu, timeout)
	if err != nil {
		return err
	}
	if resp.IsError {
		return fmt.Errorf("modbus write error: %s", resp.ErrorMsg)
	}
	return nil
}

// poll 轮询 deviceId 所有点位（TCP / RTU 共用）
//
// timeout 是**整体预算**：按剩余时间分摊给每个 poller，
// 而不是每个 poller 各自独占一份 timeout（否则设备离线时总耗时 = N × timeout）。
func (b *baseProtocol) poll(tr transport, deviceId string, timeout time.Duration) (map[string]any, error) {
	dev, model, err := b.resolveModel(deviceId)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(timeout)
	results := make(map[string]any)
	for _, p := range model.Pollers {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, protocol.ErrTimeout
		}
		resp, err := b.transactRead(tr, dev.Slave, p.FC(),
			BuildReadRequest(p.StartAddr(), p.Quantity()), p.StartAddr(), p.Quantity(), remaining)
		if err != nil {
			return nil, err
		}
		if resp.IsError {
			return nil, fmt.Errorf("modbus response error: %s", resp.ErrorMsg)
		}
		vals, err := model.DecodeResponse(dev.Slave, dev.Slave, resp)
		if err != nil {
			return nil, err
		}
		for k, v := range vals {
			results[k] = v
		}
	}

	return results, nil
}

// =====================================================================
// 辅助函数
// =====================================================================

// effectiveTimeout 返回最终生效的超时值
func effectiveTimeout(defaultTimeout, timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	}
	return defaultTimeout
}

// newModbusNormalizer 返回 Modbus 物模型归一化函数
// proto 是协议名（"tcp" 或 "rtu"），闭包捕获后 Normalizer 内部直接用
func newModbusNormalizer(proto string) protocol.Normalizer {
	return func(model *thing.Model) (any, error) {
		return NewModel("", proto, model.Properties, model.Pollers)
	}
}

// modbusTCPManifest Modbus TCP 协议注册（含 Factory + Normalizer）
var modbusTCPManifest = protocol.Manifest{
	Name:    "Modbus TCP",
	Version: "1.0",
	Registers: []protocol.Register{
		{Name: "线圈", Value: "coil"},
		{Name: "离散输入", Value: "discrete_input"},
		{Name: "保持寄存器", Value: "holding_register"},
		{Name: "输入寄存器", Value: "input_register"},
	},
	Formats: []protocol.Format{
		{Name: "uint8", Value: "uint8"},
		{Name: "int8", Value: "int8"},
		{Name: "uint16", Value: "uint16"},
		{Name: "int16", Value: "int16"},
		{Name: "uint32", Value: "uint32"},
		{Name: "int32", Value: "int32"},
		{Name: "float32", Value: "float32"},
		{Name: "float64", Value: "float64"},
		{Name: "bool", Value: "bool"},
		{Name: "bcd", Value: "bcd"},
	},
	Normalizer: newModbusNormalizer(protocol.TCP),
}

// modbusRTUManifest Modbus RTU 协议注册（含 Factory + Normalizer）
var modbusRTUManifest = protocol.Manifest{
	Name:    "Modbus RTU",
	Version: "1.0",
	Registers: []protocol.Register{
		{Name: "线圈", Value: "coil"},
		{Name: "离散输入", Value: "discrete_input"},
		{Name: "保持寄存器", Value: "holding_register"},
		{Name: "输入寄存器", Value: "input_register"},
	},
	Formats: []protocol.Format{
		{Name: "uint8", Value: "uint8"},
		{Name: "int8", Value: "int8"},
		{Name: "uint16", Value: "uint16"},
		{Name: "int16", Value: "int16"},
		{Name: "uint32", Value: "uint32"},
		{Name: "int32", Value: "int32"},
		{Name: "float32", Value: "float32"},
		{Name: "float64", Value: "float64"},
		{Name: "bool", Value: "bool"},
		{Name: "bcd", Value: "bcd"},
	},
	Normalizer: newModbusNormalizer(protocol.RTU),
}
