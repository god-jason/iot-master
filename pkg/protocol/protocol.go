// Package protocol 物联网协议统一抽象层
//
// 架构：
//
//	pkg/protocol/         共享接口层（本包）
//	├── protocol.go       Protocol 接口 + 公共类型 + 工厂注册
//	├── model.go          全局物模型缓存（RegisterModel/GetModel/UnregisterModel）
//	├── modbus/           Modbus TCP / RTU
//	├── cjt188/           建设部水表
//	├── dlt645/           国网电度表
//	├── jt808/            部标 808 车载终端
//	└── iec/              IEC 60870-5-101/103/104
//
//	internal/gateway/     网关框架，import 本包
//
// 依赖方向：gateway → protocol 基础包，protocol 子包 → protocol 基础包。
// 零循环依赖。
package protocol

import (
	"fmt"
	"io"
	"time"

	"github.com/god-jason/iot-master/pkg/thing"
)

// ============================================================
// 协议类型常量（直接 string，不再定义 type Protocol string）
// ============================================================

const (
	TCP    = "tcp"    // Modbus TCP（MBAP）
	RTU    = "rtu"    // Modbus RTU（串口帧格式 over TCP）
	CJT188 = "cjt188" // 建设部水表协议
	DLT645 = "dlt645" // 国网电度表协议
	JT808  = "jt808"  // 部标 808-2019 车载终端
	IEC101 = "iec101" // IEC 60870-5-101 串口版
	IEC103 = "iec103" // IEC 60870-5-103 保护设备
	IEC104 = "iec104" // IEC 60870-5-104 TCP 版
)

// ============================================================
// 错误定义
// ============================================================

var (
	ErrNeedMore     = fmt.Errorf("protocol: 数据不完整，等待后续")
	ErrCRC          = fmt.Errorf("protocol: CRC/校验失败")
	ErrInvalidFrame = fmt.Errorf("protocol: 无效帧")
	ErrUnknown      = fmt.Errorf("protocol: 未知协议")
	ErrNotMounted   = fmt.Errorf("protocol: 设备未挂载")
	ErrTimeout      = fmt.Errorf("protocol: 请求超时")
)

type OnValues func(device string, values map[string]any)

// ============================================================
// Protocol —— 有状态 per-connection 协议接口
//
// 每个 TCP 连接一个独立 Protocol 实例（通过 NewProtocol() 工厂创建）。
// 物模型由外部通过 RegisterModel 全局注册，Protocol 实例通过 GetModel 按需读取。
//
// 生命周期：
//
//	protocol.RegisterModel("m1", &protocol.Model{...})  ← 外部维护，一次注册全局共享
//	NewProtocol("tcp")  → Attach(conn)
//	→ Mount(devId, addr, modelId)  → Poll/Read/Write  → OnData(buf) (处理字节流)
//	→ UnMount(devId)  → Close
//
// 关键约束：OnData 被 gnet OnTraffic 调用（event loop 线程），不能阻塞。
// 跨帧拼包由调用层负责（调用层维护缓冲，循环调用 OnData 直到 consumed==0）。
// Poll/Read/Write 返回语义值 map[string]any，内部必须有异步请求-响应关联机制
// （pending request map + channel / wait group）。
// ============================================================

type Protocol interface {
	// Name 返回协议标识（应等于 Protocol 常量之一）
	Name() string

	// Attach 绑定到底层连接（发送数据用）
	// 传入的是标准 net.Conn — 实现内部按需要 wrap
	Attach(link io.ReadWriteCloser, onValues OnValues) error

	// Close 关闭协议会话（释放资源：pending channel、goroutine、内部缓冲）
	Close() error

	// Mount 挂载子设备 — deviceId 是子设备唯一标识，address 是协议地址（如 Modbus slaveAddr）
	// modelId 是该设备绑定的物模型 ID（必须已通过 protocol.RegisterModel 注册）
	// 同一个 Protocol 可以 Mount 多个 deviceId
	Mount(deviceId string, address string, modelId string) error

	// UnMount 卸载子设备
	UnMount(deviceId string)

	// Poll 轮询一个子设备所有已注册点位，返回语义值
	Poll(deviceId string, timeout time.Duration) (values map[string]any, err error)

	// Read 读取子设备某个点位的当前值
	Read(deviceId string, key string, timeout time.Duration) (value any, err error)

	// Write 写一个值到子设备某个点位
	Write(deviceId string, key string, value any, timeout time.Duration) error

	// OnData 处理一段字节流，返回本次消费的字节数。
	// 拼包（跨帧累积）由调用层负责：调用层维护缓冲区，循环调用 OnData 直到 consumed==0。
	// consumed==0 表示数据不足需等待更多字节；err 仅用于诊断日志，不影响消费循环。
	// —— 被 gnet OnTraffic 调用，不能阻塞
	OnData(buf []byte) (consumed int, err error)
}

// ============================================================
// Manifest —— 协议注册（子包 init() 里用）
//
// 包含：协议描述（供前端表单渲染）+ 工厂函数 + 物模型解析器
// 子包 init() 里调 RegisterProtocol 时传入完整 Manifest。
// ============================================================

// Register 协议寄存器定义（前端表单渲染用）
type Register struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// Format 协议数据格式定义（前端表单渲染用）
type Format struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// Factory 协议工厂函数，返回新的有状态 Protocol 实例
type Factory func() Protocol

// Normalizer 物模型归一化 —— 把 DB 原始 thing.Model 解析成协议专属模型
// 子包 init() 里填入自身的归一化函数（如 Modbus NewModel）
type Normalizer func(model *thing.Model) (any, error)

// Manifest 协议完整注册信息
type Manifest struct {
	Name       string     `json:"name,omitempty"`
	Version    string     `json:"version,omitempty"`
	Registers  []Register `json:"registers,omitempty"`
	Formats    []Format   `json:"formats,omitempty"`
	Factory    Factory    `json:"-"`
	Normalizer Normalizer `json:"-"`
}

// ============================================================
// 工厂注册
// ============================================================

var protocolRegistry = map[string]Manifest{}

// RegisterProtocol 注册协议
// 子包 init() 里调用。ma.Factory 和 ma.Parser 必须填写。
func RegisterProtocol(name string, ma Manifest) {
	protocolRegistry[name] = ma
}

// NewProtocol 根据协议名创建新的 Protocol 实例
// 每个 TCP 连接应该调一次 NewProtocol()，拿到独立的有状态实例
func NewProtocol(name string) (Protocol, bool) {
	ma, ok := protocolRegistry[name]
	if !ok {
		return nil, false
	}
	return ma.Factory(), true
}

// GetManifest 获取协议注册信息（含 Parser）
func GetManifest(name string) Manifest {
	return protocolRegistry[name]
}

// Protocols 获取所有协议注册信息（含描述）
func Protocols() []Manifest {
	var ma []Manifest
	for _, d := range protocolRegistry {
		ma = append(ma, d)
	}
	return ma
}
