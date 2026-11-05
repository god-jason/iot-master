package broker

// 主题、Topic 构造与消息结构统一复用 internal/mqtt，避免重复定义。

import "github.com/god-jason/iot-master/internal/mqtt"

// 主题常量别名：与 internal/mqtt 保持同名同值，broker 内部直接按后缀分发上行消息。
const (
	TopicPrefix = mqtt.TopicPrefix

	PropertyReport     = mqtt.PropertyReport
	PropertyReportResp = mqtt.PropertyReportResp
	PropertySet        = mqtt.PropertySet
	PropertySetResp    = mqtt.PropertySetResp
	PropertyGet        = mqtt.PropertyGet
	PropertyGetResp    = mqtt.PropertyGetResp

	ServiceCall     = mqtt.ServiceCall
	ServiceCallResp = mqtt.ServiceCallResp

	SettingSet     = mqtt.SettingSet
	SettingSetResp = mqtt.SettingSetResp
	SettingGet     = mqtt.SettingGet
	SettingGetResp = mqtt.SettingGetResp

	DatabaseSync     = mqtt.DatabaseSync
	DatabaseSyncResp = mqtt.DatabaseSyncResp
	DatabaseLoad     = mqtt.DatabaseLoad
	DatabaseLoadResp = mqtt.DatabaseLoadResp

	EventReport     = mqtt.EventReport
	EventReportResp = mqtt.EventReportResp

	StatusReport     = mqtt.StatusReport
	StatusReportResp = mqtt.StatusReportResp

	OTAReport     = mqtt.OTAReport
	OTAReportResp = mqtt.OTAReportResp
)

// BuildTopic 构造 MQTT Topic
func BuildTopic(id, suffix string) string { return mqtt.BuildTopic(id, suffix) }

// —— 消息 Payload 结构别名 ——

// 以下类型均为 internal/mqtt 中同名类型的别名，保证平台与设备两端的 Payload 结构一致。
type (
	PropertyReportMsg     = mqtt.PropertyReportMsg
	PropertySetMsg        = mqtt.PropertySetMsg
	PropertyGetMsg        = mqtt.PropertyGetMsg
	EventReportMsg        = mqtt.EventReportMsg
	StatusReportMsg       = mqtt.StatusReportMsg
	OTAReportMsg          = mqtt.OTAReportMsg
	ServiceCallMsg        = mqtt.ServiceCallMsg
	SettingSetMsg         = mqtt.SettingSetMsg
	SettingGetMsg         = mqtt.SettingGetMsg
	DatabaseSyncMsg       = mqtt.DatabaseSyncMsg
	DatabaseSyncParams    = mqtt.DatabaseSyncParams
	DatabaseLoadMsg       = mqtt.DatabaseLoadMsg
	DatabaseLoadParams    = mqtt.DatabaseLoadParams
	ResponseMessage       = mqtt.ResponseMessage
	ReportResponseMessage = mqtt.ReportResponseMessage
)
