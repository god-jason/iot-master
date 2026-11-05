// Package iot — 设备运行时统一入口。
//
// 三种接入类型各自实现 Device interface 并注册到 registry：
//   - internal/broker  — Device（内嵌 gnet broker 服务端，请求-响应模式）
//   - internal/mqtt    — Device（外置 broker 客户端，fire-and-forget）
//   - internal/gateway — Device（虚拟网关，Poll 有真实实现）
//
// broker/mqtt/gateway 监测到设备上线时构造对应 Device 实现，调 iot.Register(deviceID, dev)；
// 设备离线时调 iot.Unregister(deviceID)。业务层通过 iot.GetDevice() 统一查询。
//
// BaseDevice 提供通用的身份/状态/入站回调实现，三种实现**组合**它，只写自己特有的出站指令。
package iot

import (
	"sync"
	"time"

	"github.com/spf13/cast"
	"xorm.io/builder"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/thing"
	"github.com/god-jason/iot-master/pkg/tsdb"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

// =====================================================================
// Device interface — 运行时设备能力契约
// =====================================================================

// Device 运行时设备能力接口。
// 所有接入类型（内嵌 broker / 外置 mqtt / Modbus 虚拟网关）必须实现。
//
// 接口分三组：
//
//	身份与状态 — ID / ModelID / IsOnline / Property / Properties / LastEvent
//	入站回调   — ReportProperties / ReportEvent / ReportStatus / SetOnline / SetOffline
//	出站指令   — SetProperty / GetProperty / Control / SetConfig / GetConfig / Poll
type Device interface {
	// ===== 身份 =====

	// ID 设备唯一标识（= DB 表里 types.Device.ID）
	ID() string

	// ModelID 物模型 ID（= DB 表里 types.Device.ModelID）
	ModelID() string

	// ===== 状态查询 =====

	// IsOnline 当前在线状态
	IsOnline() bool

	// Property 单个属性最近值
	Property(name string) (interface{}, bool)

	// Properties 属性值快照（拷贝）
	Properties() map[string]interface{}

	// LastEvent 最近一次设备事件
	LastEvent() (name string, at time.Time)

	// ===== 入站回调（Transport 层调）=====

	// ReportProperties 属性上报
	ReportProperties(props map[string]interface{}, timestamp time.Time) error

	// ReportEvent 事件上报
	ReportEvent(name string, params map[string]interface{}, timestamp time.Time) error

	// ReportStatus 状态上报（电量/固件/定位等）
	ReportStatus(params map[string]interface{}, timestamp time.Time) error

	// SetOnline 标记设备上线
	SetOnline()

	// SetOffline 标记设备离线
	SetOffline()

	// ===== 出站指令（业务层调）=====

	// SetProperty 下发属性（需要设备 ACK）
	SetProperty(props map[string]interface{}) error

	// GetProperty 读取属性（需要设备 ACK）
	GetProperty(names []string) (map[string]interface{}, error)

	// Control 调用设备服务（需要设备 ACK）
	Control(service string, params map[string]interface{}) (map[string]interface{}, error)

	// SetConfig 下发设备配置
	SetConfig(data map[string]interface{}) error

	// GetConfig 读取设备配置
	GetConfig(keys []string) (map[string]interface{}, error)

	// Poll 主动采集（只有 Modbus 轮询型实现需要，MQTT 接入返回 ErrNotSupported）
	Poll() (map[string]interface{}, error)
}

// =====================================================================
// BaseDevice — 通用运行时基础实现（实现 Device 的身份/状态/入站部分）
// =====================================================================

// BaseDevice 设备运行时基础实现。
// 持有 DB 实体 + 实时状态（属性最近值 / 在线状态 / 最近事件），
// 实现 Device interface 中**通用**的部分（身份 / 状态查询 / 入站回调）。
//
// 三种 Transport 实现**组合** BaseDevice，只写自己特有的**出站指令**：
//
//	type BrokerDevice struct {
//	    *iot.BaseDevice               // 通用状态 + 入站
//	    svc *BrokerService            // transport 特有的出站
//	    ...
//	}
type BaseDevice struct {
	db *types.Device // DB 实体（身份：ID / ModelID）

	mu        sync.RWMutex
	props     map[string]interface{} // 属性最近值
	online    bool                   // 当前在线
	onlineAt  time.Time              // 最近上线时间
	offlineAt time.Time              // 最近下线时间
	eventName string                 // 最近事件名
	eventAt   time.Time              // 最近事件时间
}

// NewBaseDevice 从 DB 实体构造基础设备
func NewBaseDevice(db *types.Device) *BaseDevice {
	return &BaseDevice{
		db:    db,
		props: make(map[string]interface{}),
	}
}

// ===== 身份 =====

// ID 设备唯一标识（取自 DB 实体）
func (b *BaseDevice) ID() string { return b.db.ID }

// ModelID 设备所属物模型 ID（取自 DB 实体）
func (b *BaseDevice) ModelID() string { return b.db.ModelID }

// ===== 状态查询 =====

// IsOnline 当前在线状态（读锁保护）
func (b *BaseDevice) IsOnline() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.online
}

// Property 单个属性最近值；ok=false 表示该属性尚未上报
func (b *BaseDevice) Property(name string) (interface{}, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.props[name]
	return v, ok
}

// Properties 返回属性快照的副本，调用方可安全读写，不会影响内部状态
func (b *BaseDevice) Properties() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	cp := make(map[string]interface{}, len(b.props))
	for k, v := range b.props {
		cp[k] = v
	}
	return cp
}

// LastEvent 最近一次设备事件名与发生时间
func (b *BaseDevice) LastEvent() (string, time.Time) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.eventName, b.eventAt
}

// ===== 入站回调 =====

// SetOnline 标记设备上线并记录上线时间
func (b *BaseDevice) SetOnline() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.online = true
	b.onlineAt = time.Now()
}

// SetOffline 标记设备离线并记录下线时间；已离线时重复调用不覆盖首次下线时间
func (b *BaseDevice) SetOffline() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.online {
		b.online = false
		b.offlineAt = time.Now()
	}
}

// ReportProperties 属性上报：物模型校验 → 存状态 → 入队批量写 tsdb
func (b *BaseDevice) ReportProperties(props map[string]interface{}, timestamp time.Time) error {
	modelID := b.ModelID()
	var model *thing.Model
	modelOK := false
	if modelID != "" {
		model, modelOK = GetModel(modelID)
	}
	if modelOK {
		// 1. 属性 Min/Max 越界（仅 Min < Max 时生效）
		// 2. Filter 任意一条不通过
		// 任一不满足则整条数据丢弃：既不更新实时状态，也不写时序库
		// 丢弃时记录被拦截的具体字段/过滤器，便于调用方与运维定位（成功路径语义不变）
		if !thing.ApplyPropertyRange(model.Properties, props) {
			for _, p := range model.Properties {
				if p.Min >= p.Max || p.Key == "" {
					continue
				}
				v, ok := props[p.Key]
				if !ok {
					continue
				}
				fv := cast.ToFloat64(v)
				if fv < p.Min || fv > p.Max {
					logger.Warn("属性上报越界，整条丢弃", "device", b.ID(), "model", modelID,
						"property", p.Key, "value", v, "min", p.Min, "max", p.Max)
				}
			}
			return nil
		}
		if !thing.ApplyFilters(model.Filters, props) {
			names := make([]string, 0, len(model.Filters))
			for _, f := range model.Filters {
				if f.Name != "" && f.Expr != "" {
					names = append(names, f.Name)
				}
			}
			logger.Warn("属性上报未通过物模型过滤器，整条丢弃", "device", b.ID(), "model", modelID, "filters", names)
			return nil
		}
	}

	b.mu.Lock()
	for k, v := range props {
		b.props[k] = v
	}
	b.mu.Unlock()

	if modelOK {
		tsdb.Write(tsdbapi.Point{
			Measurement: modelID,
			Tags:        map[string]string{"id": b.ID()},
			Fields:      props,
			Time:        timestamp,
		})
	}

	return nil
}

// ReportEvent 事件上报：存状态
func (b *BaseDevice) ReportEvent(name string, params map[string]interface{}, timestamp time.Time) error {
	b.mu.Lock()
	b.eventName = name
	b.eventAt = time.Now()
	b.mu.Unlock()
	return nil
}

// ReportStatus 状态上报（电量/固件/定位等）
// 当前主要处理固件版本：状态里带 firmware 字段且与已记录版本不同时，异步回写 device 表，
// 这样设备主动上报版本或 OTA 重启后上报新版本都能及时反映到设备列表。
func (b *BaseDevice) ReportStatus(params map[string]interface{}, timestamp time.Time) error {
	raw, ok := params["firmware"]
	if !ok {
		return nil
	}
	version := cast.ToString(raw)
	if version == "" {
		return nil
	}

	b.mu.Lock()
	b.props["firmware"] = version
	b.mu.Unlock()

	// 版本未变化则跳过 DB 写入，避免设备高频上报造成无谓更新
	b.mu.RLock()
	changed := version != b.db.FirmwareVersion
	b.mu.RUnlock()
	if !changed {
		return nil
	}

	b.mu.Lock()
	b.db.FirmwareVersion = version
	b.db.FirmwareUpdatedAt = time.Now()
	b.mu.Unlock()

	deviceID := b.ID()
	go func() {
		_, err := db.UpdateOne("device", builder.Eq{"id": deviceID}, map[string]interface{}{
			"firmware_version":    version,
			"firmware_updated_at": time.Now(),
		})
		if err != nil {
			logger.Warn("固件版本回写失败", "device", deviceID, "version", version, "error", err)
		}
	}()
	return nil
}
