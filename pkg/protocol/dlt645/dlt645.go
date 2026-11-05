// Package dlt645 国家电网 DL/T 645-2007 多功能电度表协议
//
// 帧结构（无转义版，对齐 Lua 参考实现）：
//
//	[ FE FE FE FE ]  前导码（可选）
//	[ 0x68 ]          SOH
//	[ ADDR 6B ]       地址，BCD，低位在前
//	[ 0x68 ]          SOH
//	[ CTRL 1B ]       控制字
//	[ LEN 1B ]        数据域长度
//	[ DATA LEN B ]    数据域（请求+0x33，响应-0x33）
//	[ CHK 1B ]        校验和 = sum(SOH..DATA) & 0xFF
//	[ 0x16 ]          EOT
package dlt645

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

// ============================================================
// 常量
// ============================================================

const (
	SOH   byte = 0x68
	EOT   byte = 0x16
	FE    byte = 0xFE // 前导码
	FECnt      = 4    // 前导码 FE 个数

	// 控制字（主站发）
	CtrlReadData   byte = 0x11 // 读数据
	CtrlReadAddr   byte = 0x13 // 读地址
	CtrlWriteData  byte = 0x14 // 写数据
	CtrlReadClock  byte = 0x12 // 读时钟
	CtrlWriteClock byte = 0x15 // 写时钟
	CtrlReadRate   byte = 0x17 // 读费率
	CtrlWriteRate  byte = 0x18 // 写费率
	CtrlPing       byte = 0x04 // 通信复位

	// 控制字（从站应答）
	CtrlSlaveNormal byte = 0x94 // 正常应答（主站 Ctrl | 0x80）
	CtrlSlaveError  byte = 0xD4 // 从站异常应答

	// 常见 DI（数据标识符）
	DIIntActiveEnergy   uint32 = 0x90100100 // 正向有功总电能
	DIIntReactiveEnergy uint32 = 0x90200100 // 正向无功总电能
	DIVoltageA          uint32 = 0x20000100 // A相电压
	DICurrentA          uint32 = 0x20010100 // A相电流
	DIPower             uint32 = 0x20020100 // 有功功率
)

// ============================================================
// 帧 codec
// ============================================================

// Frame DLT645 帧
type Frame struct {
	Addr [6]byte // 表地址（低位在前存储，发送时也直接用）
	Ctrl byte    // 控制字
	Data []byte  // 数据域（原始，未 +/-0x33）
}

// sumChecksum 求字节数组累加和 & 0xFF
func sumChecksum(b []byte) byte {
	var s byte
	for _, v := range b {
		s += v
	}
	return s
}

// add33 数据域每个字节 +0x33（请求前编码）
func add33(b []byte) []byte {
	out := make([]byte, len(b))
	for i, v := range b {
		out[i] = (v + 0x33) & 0xFF
	}
	return out
}

// sub33 数据域每个字节 -0x33（响应后解码）
func sub33(b []byte) []byte {
	out := make([]byte, len(b))
	for i, v := range b {
		out[i] = (v - 0x33) & 0xFF
	}
	return out
}

// Encode 编码成完整报文（含 FEFEFEFE 前导码）
// DLT645 协议 DATA 域线上传输统一 +0x33
// 使用者传入的 f.Data 应该是原始（未 +0x33）状态
func (f *Frame) Encode() []byte {
	data := add33(f.Data)

	// 构建帧（不含前导码）
	lenByte := byte(len(data))
	payload := make([]byte, 0, 2+6+1+1+len(data))
	payload = append(payload, SOH)
	payload = append(payload, f.Addr[:]...)
	payload = append(payload, SOH)
	payload = append(payload, f.Ctrl)
	payload = append(payload, lenByte)
	payload = append(payload, data...)

	chk := sumChecksum(payload)

	frame := make([]byte, 0, FECnt+len(payload)+2)
	// 前导码
	for i := 0; i < FECnt; i++ {
		frame = append(frame, FE)
	}
	frame = append(frame, payload...)
	frame = append(frame, chk, EOT)
	return frame
}

// Decode 从字节流里切一帧出来，返回 Frame、剩余数据、错误
// 对齐 Lua ask() 中响应处理：
//  1. 跳过前导 FE（0~4 个）
//  2. 找 SOH
//  3. 从 len 字段算帧长度
//  4. 校验 + sub33
func Decode(data []byte) (*Frame, []byte, error) {
	// 找到 SOH（跳过前导 FE）
	sohIdx := -1
	for i := 0; i < len(data); i++ {
		if data[i] == FE {
			continue
		}
		if data[i] == SOH {
			sohIdx = i
			break
		}
		// 非 FE 也非 SOH → 跳过
		return nil, nil, fmt.Errorf("dlt645: 非法起始字节 0x%02X", data[i])
	}
	if sohIdx == -1 {
		return nil, nil, protocol.ErrNeedMore
	}

	// 最少帧：SOH(1)+ADDR(6)+SOH(1)+CTRL(1)+LEN(1)+DATA(0)+CHK(1)+EOT(1) = 12
	const headerLen = 10 // SOH+ADDR+SOH+CTRL+LEN
	if len(data)-sohIdx < headerLen {
		return nil, nil, protocol.ErrNeedMore
	}

	off := sohIdx
	// 第二 SOH 校验
	if data[off+7] != SOH {
		return nil, nil, fmt.Errorf("dlt645: 第二 SOH 错误 0x%02X", data[off+7])
	}

	frame := &Frame{}
	copy(frame.Addr[:], data[off+1:off+7])
	frame.Ctrl = data[off+8]
	length := int(data[off+9])

	totalLen := headerLen + length + 2 // +CHK +EOT
	if len(data)-off < totalLen {
		return nil, nil, protocol.ErrNeedMore
	}

	payload := data[off : off+totalLen]
	chkExpected := payload[len(payload)-2]
	chkActual := sumChecksum(payload[:len(payload)-2])
	if chkExpected != chkActual {
		return nil, nil, fmt.Errorf("dlt645: 校验错误 want=0x%02X got=0x%02X", chkExpected, chkActual)
	}
	if payload[len(payload)-1] != EOT {
		return nil, nil, fmt.Errorf("dlt645: EOT 错误 0x%02X", payload[len(payload)-1])
	}

	// 提取 DATA（原始 +0x33 状态 → sub33）
	rawData := payload[headerLen : headerLen+length]
	frame.Data = sub33(rawData)

	return frame, data[off+totalLen:], nil
}

// ============================================================
// Protocol
// ============================================================

// Protocol DLT645 协议实例
type Protocol struct {
	link     io.ReadWriteCloser
	onValues protocol.OnValues
	mu       sync.Mutex

	mounted    sync.Map // deviceId → [6]byte addr
	mountModel sync.Map // deviceId → modelId

	pendingMu sync.Mutex
	pending   map[string]*pendingRequest

	timeout time.Duration
}

type dltCachedModel struct {
	mappers []*DLT645Mapper
}

type pendingRequest struct {
	key string
	ch  chan *Frame
}

func NewProtocol() *Protocol {
	return &Protocol{
		pending: make(map[string]*pendingRequest),
		timeout: 3 * time.Second,
	}
}

func (*Protocol) Name() string { return protocol.DLT645 }

func (c *Protocol) Attach(link io.ReadWriteCloser, onValues protocol.OnValues) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.link = link
	c.onValues = onValues
	return nil
}

func (c *Protocol) Close() error {
	c.pendingMu.Lock()
	for _, p := range c.pending {
		close(p.ch)
	}
	c.pending = make(map[string]*pendingRequest)
	c.pendingMu.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mounted = sync.Map{}
	c.mountModel = sync.Map{}
	c.link = nil
	return nil
}

func (c *Protocol) Mount(deviceId string, address string, modelId string) error {
	addr, err := parse6ByteAddr(address)
	if err != nil {
		return err
	}
	c.mounted.Store(deviceId, addr)
	c.mountModel.Store(deviceId, modelId)
	return nil
}

func (c *Protocol) UnMount(deviceId string) {
	c.mounted.Delete(deviceId)
	c.mountModel.Delete(deviceId)
}

func (c *Protocol) Poll(deviceId string, timeout time.Duration) (map[string]any, error) {
	to := c.timeout
	if timeout > 0 {
		to = timeout
	}

	addrAny, ok := c.mounted.Load(deviceId)
	if !ok {
		return nil, protocol.ErrNotMounted
	}
	addr := addrAny.([6]byte)

	modelIdAny, _ := c.mountModel.Load(deviceId)
	cached, ok := c.findCached(modelIdAny.(string))
	if !ok {
		return nil, fmt.Errorf("dlt645: 设备 %s 未注册物模型", deviceId)
	}

	results := make(map[string]any)
	for _, mapper := range cached.mappers {
		reqDatas := mapper.EncodeRequest()
		for _, reqData := range reqDatas {
			frame := &Frame{Addr: addr, Ctrl: CtrlReadData, Data: reqData}
			key := pendingKey(addr, CtrlReadData)

			c.pendingMu.Lock()
			ch := make(chan *Frame, 1)
			c.pending[key] = &pendingRequest{key: key, ch: ch}
			c.pendingMu.Unlock()

			c.mu.Lock()
			var err error
			if c.link != nil {
				_, err = c.link.Write(frame.Encode())
			} else {
				err = fmt.Errorf("dlt645: 未 Attach 连接")
			}
			c.mu.Unlock()

			if err != nil {
				c.pendingMu.Lock()
				delete(c.pending, key)
				c.pendingMu.Unlock()
				return nil, err
			}

			select {
			case resp := <-ch:
				c.pendingMu.Lock()
				delete(c.pending, key)
				c.pendingMu.Unlock()
				if resp == nil {
					return nil, protocol.ErrTimeout
				}
				if resp.Ctrl == CtrlSlaveError {
					return nil, fmt.Errorf("dlt645: 从站异常应答")
				}
				vals, err := mapper.DecodeResponse(resp.Ctrl, resp.Data)
				if err != nil {
					return nil, err
				}
				for k, v := range vals {
					results[k] = v
				}
			case <-time.After(to):
				c.pendingMu.Lock()
				delete(c.pending, key)
				c.pendingMu.Unlock()
				return nil, protocol.ErrTimeout
			}
		}
	}
	return results, nil
}

func (c *Protocol) Read(deviceId string, key string, timeout time.Duration) (any, error) {
	vals, err := c.Poll(deviceId, timeout)
	if err != nil {
		return nil, err
	}
	if v, ok := vals[key]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("dlt645: key %q 不存在", key)
}

func (c *Protocol) Write(deviceId string, key string, value any, timeout time.Duration) error {
	return fmt.Errorf("dlt645 Write: 暂未实现")
}

// OnData 处理一段字节流，返回消费字节数（跨帧拼包由调用层负责）。
func (c *Protocol) OnData(buf []byte) (int, error) {
	frame, remaining, err := Decode(buf)
	switch err {
	case nil:
		frameLen := len(buf) - len(remaining)

		c.pendingMu.Lock()
		key := pendingKey(frame.Addr, CtrlReadData)
		if p, ok := c.pending[key]; ok {
			select {
			case p.ch <- frame:
			default:
			}
			delete(c.pending, key)
		}
		c.pendingMu.Unlock()
		return frameLen, nil

	case protocol.ErrNeedMore:
		return 0, nil
	default:
		// 非法起始字节等错误，跳过 1 字节交由调用层重试
		return 1, nil
	}
}

func pendingKey(addr [6]byte, ctrl byte) string {
	return fmt.Sprintf("%x:%02X", addr[:], ctrl)
}

func (c *Protocol) findCached(modelId string) (*dltCachedModel, bool) {
	v, ok := protocol.GetModel(modelId)
	if !ok {
		return nil, false
	}
	m, ok := v.(*dltCachedModel)
	return m, ok
}

// ============================================================
// 地址解析
// ============================================================

// parse6ByteAddr 解析表地址 → [6]byte（低位在前存储）
// 支持: "a.b.c.d.e.f" 点分 6 段 / 完整 hex / 纯数字
func parse6ByteAddr(s string) ([6]byte, error) {
	var out [6]byte
	s = strings.TrimSpace(s)
	if s == "" {
		return out, fmt.Errorf("地址为空")
	}
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if len(parts) > 6 {
			return out, fmt.Errorf("地址段数 > 6")
		}
		// 高位在前写，存储时低位在前 → 逆序
		tmp := make([]byte, len(parts))
		for i, p := range parts {
			v, err := parseUint8(strings.TrimSpace(p))
			if err != nil {
				return out, fmt.Errorf("地址段 %q: %w", p, err)
			}
			tmp[i] = v
		}
		// 逆序 → 低位在前存储
		for i := range tmp {
			out[5-i] = tmp[i]
		}
		return out, nil
	}
	var v uint64
	if _, err := fmt.Sscanf(s, "%x", &v); err != nil {
		if _, err = fmt.Sscanf(s, "%d", &v); err != nil {
			return out, fmt.Errorf("地址 %q 无法解析", s)
		}
	}
	for i := 5; i >= 0 && v > 0; i-- {
		out[i] = byte(v & 0xFF)
		v >>= 8
	}
	return out, nil
}

func parseUint8(s string) (byte, error) {
	var v uint64
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		if _, err = fmt.Sscanf(s, "%x", &v); err != nil {
			return 0, err
		}
	}
	if v > 0xFF {
		return 0, fmt.Errorf("超出 uint8 范围")
	}
	return byte(v), nil
}

// 保留：binary.LittleEndian 引用
var _ = binary.LittleEndian

func init() {
	m := protocol.Manifest{
		Name:    "DLT645",
		Version: "2007",
		Registers: []protocol.Register{
			{Name: "数据标识符 DI", Value: "DI"},
		},
		Formats: []protocol.Format{
			{Name: "BCD", Value: "bcd"},
			{Name: "uint16", Value: "uint16"},
			{Name: "uint32", Value: "uint32"},
			{Name: "float32", Value: "float32"},
			{Name: "string", Value: "string"},
		},
		Normalizer: func(model *thing.Model) (any, error) {
			mapper, err := NewMapper(model.Properties)
			if err != nil {
				return nil, err
			}
			return &dltCachedModel{mappers: []*DLT645Mapper{mapper}}, nil
		},
	}
	m.Factory = func() protocol.Protocol { return NewProtocol() }
	protocol.RegisterProtocol(protocol.DLT645, m)
}

var _ protocol.Protocol = (*Protocol)(nil)
