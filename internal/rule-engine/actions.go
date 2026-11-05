package ruleengine

import (
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/tsdb"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

// actionExecutor 规则动作执行器，将命中规则的动作落到设备事件、设备控制或时序库
type actionExecutor struct {
	caller DeviceAdapter
}

// newActionExecutor 构造动作执行器，caller 提供设备控制与物模型查询能力。
func newActionExecutor(caller DeviceAdapter) *actionExecutor {
	return &actionExecutor{caller: caller}
}

// execute 顺序执行一条规则的全部动作；单个动作失败不影响其它动作
func (a *actionExecutor) execute(rule types.Rule, deviceID string, values map[string]interface{}) []types.RuleActionResult {
	results := make([]types.RuleActionResult, 0, len(rule.Actions))
	for _, act := range rule.Actions {
		results = append(results, a.runOne(rule, deviceID, values, act))
	}
	return results
}

// runOne 执行单个动作
func (a *actionExecutor) runOne(rule types.Rule, deviceID string, values map[string]interface{}, act types.RuleAction) (r types.RuleActionResult) {
	r.Type = act.Type
	// 兜底 recover：单个动作 panic 只标记本动作失败，避免影响同一规则其它动作
	defer func() {
		if rec := recover(); rec != nil {
			r.Success = false
			r.Message = "动作执行异常"
			logger.Error("规则动作执行异常", "ruleID", rule.ID, "action", act.Type, "panic", rec)
		}
	}()

	switch act.Type {
	case types.RuleActionAlarm, types.RuleActionNotify:
		// 报警/通知都落 device_event：通知固定 info 级别，报警取配置值、缺省 alert
		level, _ := act.Params["level"].(string)
		message, _ := act.Params["message"].(string)
		if act.Type == types.RuleActionNotify {
			level = "info"
		}
		if level == "" {
			level = "alert"
		}
		if message == "" {
			message = rule.Name
		}
		if err := a.writeDeviceEvent(deviceID, level, rule.Name, message, values, rule.ID); err != nil {
			r.Success = false
			r.Message = err.Error()
			return
		}
		r.Success = true
		r.Message = "已写入事件"

	case types.RuleActionControlProperty:
		props, _ := act.Params["properties"].(map[string]interface{})
		if len(props) == 0 {
			r.Success, r.Message = false, "未配置下发属性"
			return
		}
		resp, err := a.caller.SetProperty(deviceID, props)
		r.Success, r.Message = types.ResolveCallResult(resp, err)

	case types.RuleActionControlService:
		serviceName, _ := act.Params["service"].(string)
		params, _ := act.Params["params"].(map[string]interface{})
		if serviceName == "" {
			r.Success, r.Message = false, "未配置服务名"
			return
		}
		resp, err := a.caller.Control(deviceID, serviceName, params)
		r.Success, r.Message = types.ResolveCallResult(resp, err)

	case types.RuleActionData:
		r.Success, r.Message = a.writeDerivedData(deviceID, act.Params, values)

	default:
		r.Success, r.Message = false, "未知动作类型: "+act.Type
	}
	return r
}

// writeDeviceEvent 写入设备事件（报警/通知）
func (a *actionExecutor) writeDeviceEvent(deviceID, level, code, message string, values map[string]interface{}, ruleID string) error {
	event := types.DeviceEvent{
		DeviceID: deviceID,
		Type:     level,
		Code:     code,
		Message:  message,
		Data: map[string]interface{}{
			"rule_id": ruleID,
			"values":  values,
		},
		Created: time.Now(),
	}
	_, err := db.InsertOne("device_event", event)
	return err
}

// writeDerivedData 将衍生指标写入设备历史时序库。
// value 未配置时按 source 字段（缺省与 field 同名）从当前属性值中取数。
func (a *actionExecutor) writeDerivedData(deviceID string, params, values map[string]interface{}) (bool, string) {
	field, _ := params["field"].(string)
	if field == "" {
		return false, "未配置衍生字段名"
	}
	value := params["value"]
	if value == nil {
		source, _ := params["source"].(string)
		if source == "" {
			source = field
		}
		value = values[source]
	}
	modelID := a.caller.ModelIDOf(deviceID)
	if modelID == "" {
		return false, "无法获取物模型ID"
	}
	tsdb.Write(tsdbapi.Point{
		Measurement: modelID,
		Tags:        map[string]string{"id": deviceID},
		Fields:      map[string]any{field: value},
		Time:        time.Now(),
	})
	return true, "已写入衍生指标"
}
