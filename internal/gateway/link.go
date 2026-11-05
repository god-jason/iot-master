package gateway

import (
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"

	"github.com/god-jason/iot-master/pkg/protocol"
)

// Link 一条 TCP 连接的运行上下文（gnet Context），内部保存该链路下的设备。
//
// 生命周期：OnOpen 创建 → 注册包解析后确定链路 ID，按 Device.LinkID 加载设备并逐一 Mount → OnClose 销毁。
// 协议实例与拼包缓冲都由 Link 持有，Server 只保存 links。
type Link struct {
	mu sync.RWMutex // 保护注册状态、设备列表与轮询调度字段

	ServerID int64             // 所属 Server 实例 ID
	ConnID   int64             // 连接序号（Server 内自增）
	Conn     gnet.Conn         // 底层 gnet 连接
	Proto    protocol.Protocol // 该连接独立的协议实例
	Protocol string            // 协议名（创建时确定）
	Remote   string            // 对端地址

	id           string    // 链路 ID（= 注册包 SN，注册前为空）
	sn           string    // 注册包中的设备序列号
	imei         string    // 注册包中的 IMEI
	name         string    // 注册包中的名称
	registeredAt time.Time // 注册完成时间

	Registered bool   // 是否已完成注册包解析（OnTraffic 据此切换解析路径）
	buf        []byte // 跨事件累积的拼包缓冲

	devices  []*Device // 该链路挂载的设备快照
	lastSeen time.Time // 链路最近活跃时间

	// 轮询调度状态
	pollCfg  *PollConfig // 链路级兜底轮询配置（为空时用 Server 默认间隔）
	lastPoll time.Time   // 上次派发轮询的时间
	busy     bool        // 是否有轮询任务正在执行（保证同链路串行）
}

// newLink 创建连接上下文
func newLink(serverID, connID int64, conn gnet.Conn, proto protocol.Protocol, protoName string) *Link {
	return &Link{
		ServerID: serverID,
		ConnID:   connID,
		Conn:     conn,
		Proto:    proto,
		Protocol: protoName,
		Remote:   conn.RemoteAddr().String(),
		lastPoll: time.Now(),
	}
}

// ProtocolInstance 返回协议实例的并发安全快照（连接关闭摘除后为 nil）
//
// Poller / 设备指令等非 event loop goroutine 必须通过该方法访问协议实例，
// 避免与 OnClose 的关闭/摘除操作并发读写 Link.Proto。
func (l *Link) ProtocolInstance() protocol.Protocol {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Proto
}

// detachProtocol 先摘除协议实例再交回调用方关闭，避免关闭后仍被在途任务取到
func (l *Link) detachProtocol() protocol.Protocol {
	l.mu.Lock()
	p := l.Proto
	l.Proto = nil
	l.mu.Unlock()
	return p
}

// ID 链路 ID（= 注册包 SN，注册前为空）
func (l *Link) ID() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.id
}

// SN 注册包中的设备序列号
func (l *Link) SN() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.sn
}

// IMEI 注册包中的 IMEI
func (l *Link) IMEI() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.imei
}

// Name 注册包中的名称
func (l *Link) Name() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.name
}

// IsRegistered 连接是否已完成注册包解析
func (l *Link) IsRegistered() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Registered
}

// RegisteredAt 注册完成时间
func (l *Link) RegisteredAt() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.registeredAt
}

// LastSeen 最近活跃时间
func (l *Link) LastSeen() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastSeen
}

// Devices 链路下设备快照
func (l *Link) Devices() []*Device {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return copyDevices(l.devices)
}

// replaceDevices 替换设备列表，返回旧列表
func (l *Link) replaceDevices(devs []*Device) []*Device {
	l.mu.Lock()
	old := l.devices
	l.devices = devs
	l.mu.Unlock()
	return old
}

// register 写入链路身份并标记注册完成
func (l *Link) register(pkt RegisterPacket) {
	id := pkt.SN
	if id == "" {
		id = pkt.IMEI
	}
	if id == "" {
		id = pkt.Device
	}

	now := time.Now()
	l.mu.Lock()
	l.id = id
	l.sn = pkt.SN
	l.imei = pkt.IMEI
	l.name = pkt.Name
	l.registeredAt = now
	l.lastSeen = now
	l.Registered = true
	l.mu.Unlock()
}

// PollConfig 返回兜底轮询配置
func (l *Link) PollConfig() *PollConfig {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.pollCfg
}

// SetPoll 设置兜底轮询配置
func (l *Link) SetPoll(p *PollConfig) {
	l.mu.Lock()
	l.pollCfg = p
	l.lastPoll = time.Now()
	l.mu.Unlock()
}

// reschedule 重置轮询时间（设备挂载变化后调用）
func (l *Link) reschedule() {
	l.mu.Lock()
	l.lastPoll = time.Now()
	l.mu.Unlock()
}

// touch 刷新连接与设备最近活跃时间
func (l *Link) touch(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lastSeen = now
	for _, d := range l.devices {
		if d != nil {
			d.markActive(now)
		}
	}
}

// setOffline 连接断开时置离线
func (l *Link) setOffline() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, d := range l.devices {
		if d != nil {
			d.markOffline()
		}
	}
}

// copyDevices 拷贝设备列表，避免调用方直接持有内部切片
func copyDevices(src []*Device) []*Device {
	if len(src) == 0 {
		return nil
	}
	out := make([]*Device, 0, len(src))
	for _, d := range src {
		out = append(out, d)
	}
	return out
}
