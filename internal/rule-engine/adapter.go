package ruleengine

import (
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

// ModelLookup 设备→物模型映射查询能力
type ModelLookup interface {
	// ModelIDOf 返回设备所属的物模型 ID（设备不存在时返回空串）
	ModelIDOf(deviceID string) string
}

// DeviceAdapter 规则引擎所需的设备侧完整能力
// 包含 types.DeviceCaller 的全部方法，外加 ModelLookup 和 LatestData
type DeviceAdapter interface {
	// IsOnline 查询设备当前在线状态
	IsOnline(deviceID string) bool
	// SetProperty 向设备下发属性（可能需要 ACK）
	SetProperty(deviceID string, params map[string]interface{}) (*types.CallResult, error)
	// Control 调用设备服务
	Control(deviceID, service string, params map[string]interface{}) (*types.CallResult, error)
	// ModelIDOf 返回设备所属物模型 ID
	ModelIDOf(deviceID string) string
	// LatestData 返回设备各属性最新值（字段名 → 时序点），定时规则求值使用
	LatestData(deviceID string) (map[string]tsdbapi.LastResult, error)
}
