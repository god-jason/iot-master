package iot

import (
	"context"
	"encoding/json"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/tsdb"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

// EngineAdapter 把 iot 设备注册表适配为规则引擎 / 场景引擎所需的设备能力接口。
// 规则引擎与场景引擎只依赖接口，不直接依赖 iot 包，避免反向导入。
type EngineAdapter struct{}

// NewEngineAdapter 创建设备能力适配器
func NewEngineAdapter() *EngineAdapter { return &EngineAdapter{} }

// IsOnline 返回设备当前在线状态
func (a *EngineAdapter) IsOnline(deviceID string) bool {
	if d, ok := GetDevice(deviceID); ok {
		return d.IsOnline()
	}
	return false
}

// SetProperty 向设备下发属性
func (a *EngineAdapter) SetProperty(deviceID string, params map[string]interface{}) (*types.CallResult, error) {
	d, ok := GetDevice(deviceID)
	if !ok {
		return &types.CallResult{Ret: 0, Message: "设备未在线"}, nil
	}
	if err := d.SetProperty(params); err != nil {
		return &types.CallResult{Ret: 0, Message: err.Error()}, nil
	}
	return &types.CallResult{Ret: 1}, nil
}

// Control 调用设备服务
func (a *EngineAdapter) Control(deviceID, service string, params map[string]interface{}) (*types.CallResult, error) {
	d, ok := GetDevice(deviceID)
	if !ok {
		return &types.CallResult{Ret: 0, Message: "设备未在线"}, nil
	}
	res, err := d.Control(service, params)
	if err != nil {
		return &types.CallResult{Ret: 0, Message: err.Error()}, nil
	}
	msg := ""
	if res != nil {
		if b, e := json.Marshal(res); e == nil {
			msg = string(b)
		}
	}
	return &types.CallResult{Ret: 1, Message: msg}, nil
}

// ModelIDOf 返回设备所属物模型 ID
func (a *EngineAdapter) ModelIDOf(deviceID string) string {
	if d, ok := GetDevice(deviceID); ok {
		return d.ModelID()
	}
	return ""
}

// LatestData 从时序库读取设备各属性最新值（定时规则求值用）
func (a *EngineAdapter) LatestData(deviceID string) (map[string]tsdbapi.LastResult, error) {
	if tsdb.DefaultStore == nil {
		return nil, nil
	}
	d, ok := GetDevice(deviceID)
	if !ok {
		return nil, nil
	}
	return tsdb.DefaultStore.Last(context.Background(), d.ModelID(), map[string]string{"id": deviceID})
}
