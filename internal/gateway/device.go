package gateway

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
)

// ErrNotSupported 此传输不支持的操作
var ErrNotSupported = errors.New("gateway transport does not support this operation")

// Device 软网关设备 —— 归属某条 Link（TCP 连接）的设备运行时。
//
// 设备点位信息（从站地址 / 功能码 / 起始地址 / 数量）由协议实例内部维护，
// DB 实体（身份 / SN / 物模型 / 地址）通过组合的 iot.BaseDevice 持有。
type Device struct {
	*iot.BaseDevice

	mu   sync.RWMutex // 保护 mounted 与在线状态字段
	link *Link        // 所属连接上下文（协议实例与设备列表的入口）

	ServerID     int64     // 所属 Server 实例 ID
	SN           string    // 设备序列号（来自 DB 设备行）
	IMEI         string    // 设备 IMEI（物联网模组场景，普通网关通常为空）
	Name         string    // 设备名称
	ConnID       int64     // 创建时所属连接的 connID
	Remote       string    // 创建时所属连接的对端地址
	Protocol     string    // 所属连接的协议名
	Online       bool      // 在线状态（读写需经 mu）
	RegisteredAt time.Time // 设备创建（挂载）时间
	LastSeen     time.Time // 最近一次活跃时间

	mounted bool // 是否已成功 Mount 到协议实例
}

// 编译时断言 Device 完整实现了 iot.Device interface
var _ iot.Device = (*Device)(nil)

// newDevice 由 DB 设备实体创建归属 link 的运行时设备
func newDevice(link *Link, dbDev *types.Device) *Device {
	now := time.Now()
	return &Device{
		BaseDevice:   iot.NewBaseDevice(dbDev),
		link:         link,
		ServerID:     link.ServerID,
		SN:           dbDev.SN,
		Name:         dbDev.Name,
		ConnID:       link.ConnID,
		Remote:       link.Remote,
		Protocol:     link.Protocol,
		Online:       true,
		RegisteredAt: now,
		LastSeen:     now,
	}
}

// Link 返回设备所属连接上下文
func (d *Device) Link() *Link { return d.link }

// DeviceID 协议实例中使用的挂载标识（= DB 设备 ID）
func (d *Device) DeviceID() string { return d.ID() }

// Mounted 是否已成功 Mount 到协议实例
func (d *Device) Mounted() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.mounted
}

// setMounted 记录协议挂载结果
func (d *Device) setMounted(v bool) {
	d.mu.Lock()
	d.mounted = v
	d.mu.Unlock()
}

// String 日志友好输出
func (d *Device) String() string {
	return fmt.Sprintf("[Device id=%s sn=%s remote=%s proto=%s]", d.ID(), d.SN, d.Remote, d.Protocol)
}

// OnlineStatus 在线状态（并发安全读取）
func (d *Device) OnlineStatus() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Online
}

// LastSeenAt 最近活跃时间（并发安全读取）
func (d *Device) LastSeenAt() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.LastSeen
}

// markActive 标记在线并刷新最近活跃时间
func (d *Device) markActive(now time.Time) {
	d.mu.Lock()
	d.Online = true
	d.LastSeen = now
	d.mu.Unlock()
}

// markOffline 标记离线
func (d *Device) markOffline() {
	d.mu.Lock()
	d.Online = false
	d.mu.Unlock()
}

// ===== 出站指令 =====

// Poll 主动采集 —— 网关 Device 独有的核心能力
func (d *Device) Poll() (map[string]interface{}, error) {
	if d.link == nil {
		return nil, fmt.Errorf("gateway: 协议实例未初始化")
	}
	proto := d.link.ProtocolInstance()
	if proto == nil {
		return nil, fmt.Errorf("gateway: 协议实例未初始化")
	}
	return proto.Poll(d.DeviceID(), 10*time.Second)
}

// SetProperty 写入属性 —— 逐个属性走 Proto.Write()
func (d *Device) SetProperty(props map[string]interface{}) error {
	if d.link == nil {
		return fmt.Errorf("gateway: 协议实例未初始化")
	}
	proto := d.link.ProtocolInstance()
	if proto == nil {
		return fmt.Errorf("gateway: 协议实例未初始化")
	}

	model, _ := iot.GetModel(d.ModelID())

	for key, value := range props {
		// 按物模型校验可写性：只读属性直接拒绝，避免下发无效指令
		if model != nil {
			for _, p := range model.Properties {
				if p.Key == key {
					if p.Mode == "r" {
						return fmt.Errorf("gateway: property %s 只读", key)
					}
					break
				}
			}
		}
		if err := proto.Write(d.DeviceID(), key, value, 5*time.Second); err != nil {
			return fmt.Errorf("gateway.Write(%s=%v): %w", key, value, err)
		}
	}
	return nil
}

// GetProperty 读取属性 —— names 为空时走 Poll 读全部
//
// 指定多个属性时逐个读取，单个属性失败不影响其它属性：
// 返回已成功读取的部分结果，并将所有失败汇总为一个错误（无失败时 error 为 nil）。
func (d *Device) GetProperty(names []string) (map[string]interface{}, error) {
	if d.link == nil {
		return nil, fmt.Errorf("gateway: 协议实例未初始化")
	}
	proto := d.link.ProtocolInstance()
	if proto == nil {
		return nil, fmt.Errorf("gateway: 协议实例未初始化")
	}

	if len(names) == 0 {
		// 未指定属性名时退化为整设备主动采集
		return d.Poll()
	}

	result := make(map[string]interface{}, len(names))
	var errs []error
	for _, name := range names {
		v, err := proto.Read(d.DeviceID(), name, 5*time.Second)
		if err != nil {
			// 记录失败并继续读取其余属性，保留部分成功结果
			errs = append(errs, fmt.Errorf("gateway.Read(%s): %w", name, err))
			continue
		}
		result[name] = v
	}
	return result, errors.Join(errs...)
}

// Control 调用设备服务 —— 轮询型协议没有此概念
func (d *Device) Control(service string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, ErrNotSupported
}

// SetConfig 下发设备配置
func (d *Device) SetConfig(data map[string]interface{}) error {
	return ErrNotSupported
}

// GetConfig 读取设备配置
func (d *Device) GetConfig(keys []string) (map[string]interface{}, error) {
	return nil, ErrNotSupported
}
