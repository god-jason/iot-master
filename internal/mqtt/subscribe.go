package mqtt

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/ota"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"xorm.io/builder"
)

// subscription 一条上行订阅：主题过滤词及其专属 handler。
// {device_id} 用通配符 + 匹配，因此每个 handler 无需再解析 suffix。
type subscription struct {
	filter  string                             // 上行主题过滤词，设备段统一用 + 通配
	handler func(topic string, payload []byte) // 该过滤词专属的处理函数
}

// subscriptions 平台订阅的上行主题，每个 filter 绑定独立 handler
var subscriptions = []subscription{
	{BuildTopic("+", PropertyReport), handlePropertyReport},
	{BuildTopic("+", PropertyReportResp), responseHandler(PropertyReportResp)},
	{BuildTopic("+", PropertySetResp), responseHandler(PropertySetResp)},
	{BuildTopic("+", PropertyGetResp), responseHandler(PropertyGetResp)},
	{BuildTopic("+", EventReport), handleEventReport},
	{BuildTopic("+", EventReportResp), responseHandler(EventReportResp)},
	{BuildTopic("+", StatusReport), handleStatusReport},
	{BuildTopic("+", StatusReportResp), responseHandler(StatusReportResp)},
	{BuildTopic("+", OTAReport), handleOTAReport},
	{BuildTopic("+", OTAReportResp), responseHandler(OTAReportResp)},
	{BuildTopic("+", ServiceCallResp), responseHandler(ServiceCallResp)},
	{BuildTopic("+", SettingSetResp), responseHandler(SettingSetResp)},
	{BuildTopic("+", SettingGetResp), responseHandler(SettingGetResp)},
	{BuildTopic("+", DatabaseSyncResp), responseHandler(DatabaseSyncResp)},
	{BuildTopic("+", DatabaseLoadResp), responseHandler(DatabaseLoadResp)},
}

// StartSubscribe 启动外置 broker 上行主题订阅（供 boot.Startup 调用）。
// 调用前需已完成 MQTT 客户端连接（BootStartup 中 Startup 之后调用）。
func StartSubscribe() {
	for _, sub := range subscriptions {
		Subscribe(sub.filter, sub.handler)
	}
	logger.Info("外置 MQTT broker 上行主题订阅已启动", "count", len(subscriptions))
}

// ===== 上报类 handler =====

// handlePropertyReport 属性上报
func handlePropertyReport(topic string, payload []byte) {
	dev := inbound(topic)
	if dev == nil {
		return
	}
	var m PropertyReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析属性上报失败", "topic", topic, "error", err)
		return
	}
	if len(m.Params) == 0 {
		return
	}
	_ = dev.ReportProperties(m.Params, msgTime(m.Time))
}

// handleEventReport 事件上报
func handleEventReport(topic string, payload []byte) {
	dev := inbound(topic)
	if dev == nil {
		return
	}
	var m EventReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析事件上报失败", "topic", topic, "error", err)
		return
	}
	if m.Event.Name == "" {
		return
	}
	_ = dev.ReportEvent(m.Event.Name, m.Event.Params, time.Now())
}

// handleStatusReport 状态上报
func handleStatusReport(topic string, payload []byte) {
	dev := inbound(topic)
	if dev == nil {
		return
	}
	var m StatusReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析状态上报失败", "topic", topic, "error", err)
		return
	}
	if len(m.Params) == 0 {
		return
	}
	_ = dev.ReportStatus(m.Params, msgTime(m.Time))
}

// handleOTAReport OTA 升级进度上报
func handleOTAReport(topic string, payload []byte) {
	deviceID, ok := deviceIDFromTopic(topic)
	if !ok {
		return
	}
	var m OTAReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析 OTA 上报失败", "topic", topic, "error", err)
		return
	}
	if m.TaskID == "" {
		logger.Warn("OTA 上报缺少 task_id", "device", deviceID)
		return
	}
	ensureDevice(deviceID)
	ota.HandleReport(deviceID, m.TaskID, m.Stage, m.Progress, m.Error, m.Version)
}

// ===== 应答类 handler =====

// responseHandler 返回某应答主题专属的 handler：按消息 id 回投给等待中的出站请求
func responseHandler(suffix string) func(topic string, payload []byte) {
	return func(topic string, payload []byte) {
		var m ResponseMessage
		if err := json.Unmarshal(payload, &m); err != nil {
			logger.Warn("解析设备应答失败", "topic", topic, "error", err)
			return
		}
		if responseManager.HandleResponse(m.ID, &m) {
			return
		}
		deviceID, _ := deviceIDFromTopic(topic)
		logger.Debug("外置 MQTT 收到未匹配的设备应答", "device", deviceID, "suffix", suffix, "id", m.ID, "ret", m.Ret)
	}
}

// ===== 公共辅助 =====

// inbound 解析 topic → deviceID，取/注册设备并标记在线，返回 Device（失败返回 nil）
func inbound(topic string) iot.Device {
	deviceID, ok := deviceIDFromTopic(topic)
	if !ok {
		return nil
	}
	dev := ensureDevice(deviceID)
	if dev == nil {
		return nil
	}
	dev.SetOnline()
	return dev
}

// deviceIDFromTopic 从 $thing/{device_id}/{suffix} 解析 device_id
func deviceIDFromTopic(topic string) (string, bool) {
	parts := strings.SplitN(topic, "/", 3)
	if len(parts) != 3 || parts[0] != TopicPrefix || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

// msgTime 将毫秒级时间戳转为 time.Time，缺省用当前时间
func msgTime(ms int64) time.Time {
	if ms > 0 {
		return time.UnixMilli(ms)
	}
	return time.Now()
}

// ensureDevice 从 registry 取设备；未注册则查 DB 构造 Device 并 Register。
// 已禁用（disabled=true）的设备不构造、不上线。
func ensureDevice(deviceID string) iot.Device {
	if dev, ok := iot.GetDevice(deviceID); ok {
		return dev
	}

	// 查询时排除已禁用设备，避免被禁用设备经上行消息自动注册上线
	var dbDev types.Device
	found, err := db.FindOne("device", builder.Eq{"id": deviceID, "disabled": false}, &dbDev)
	if err != nil {
		logger.Warn("外置 MQTT 消息：查 DB 失败", "device", deviceID, "error", err)
		return nil
	}
	if !found || dbDev.ID == "" {
		// 再查一次区分「不存在」与「已禁用」，只对已禁用设备告警
		var raw types.Device
		if ok, findErr := db.FindOne("device", builder.Eq{"id": deviceID}, &raw); findErr == nil && ok && raw.Disabled {
			logger.Warn("外置 MQTT 消息：设备已被禁用，忽略", "device", deviceID)
		} else {
			logger.Debug("外置 MQTT 消息：设备未注册（DB 里没找到）", "device", deviceID)
		}
		return nil
	}

	dev := NewDevice(&dbDev)
	dev.SetOnline()
	iot.Register(deviceID, dev)
	logger.Info("外置 MQTT 设备已注册", "device", deviceID)
	return dev
}
