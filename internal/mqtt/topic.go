package mqtt

import "fmt"

// 主题常量：所有主题统一以 $thing 为前缀，结构为 $thing/{device_id}/{operation}。
// 下行（平台 → 设备）与上行（设备 → 平台）共用这一套后缀定义。
const (
	// TopicPrefix 所有主题的统一前缀
	TopicPrefix = "$thing"

	// 属性：上行 report（设备 → 平台）；下行 set/get（平台 → 设备）；*Resp 为对应应答
	PropertyReport     = "property/report"
	PropertyReportResp = "property/report/response"
	PropertySet        = "property/set"
	PropertySetResp    = "property/set/response"
	PropertyGet        = "property/get"
	PropertyGetResp    = "property/get/response"

	// 服务调用：下行 call（平台 → 设备），Resp 为设备应答
	ServiceCall     = "service/call"
	ServiceCallResp = "service/call/response"

	// 配置：下行 set/get（平台 → 设备），Resp 为设备应答
	SettingSet     = "setting/set"
	SettingSetResp = "setting/set/response"
	SettingGet     = "setting/get"
	SettingGetResp = "setting/get/response"

	// 数据库：下行 sync/load（平台 → 设备），Resp 为设备应答
	DatabaseSync     = "database/sync"
	DatabaseSyncResp = "database/sync/response"
	DatabaseLoad     = "database/load"
	DatabaseLoadResp = "database/load/response"

	// 事件：上行 report（设备 → 平台），Resp 为平台应答
	EventReport     = "event/report"
	EventReportResp = "event/report/response"

	// 状态：上行 report（设备 → 平台），Resp 为平台应答
	StatusReport     = "status/report"
	StatusReportResp = "status/report/response"

	// OTA 升级：上行 report（设备 → 平台，上报升级阶段与进度），Resp 为平台应答；
	// 下行指令复用 ServiceCall（service = ota_update / ota_cancel）
	OTAReport     = "ota/report"
	OTAReportResp = "ota/report/response"
)

// BuildTopic 构造 MQTT Topic
// 格式为 $thing/{id}/{suffix}；id 为设备 ID，订阅时可用 + 通配全部设备。
func BuildTopic(id, suffix string) string {
	return fmt.Sprintf("%s/%s/%s", TopicPrefix, id, suffix)
}

// —— 消息 Payload 结构定义 ——

// PropertyReportMsg 属性上报（设备 → 平台）
type PropertyReportMsg struct {
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"params"`
	Time   int64                  `json:"time"`
}

// PropertySetMsg 属性设置（平台 → 设备）
type PropertySetMsg struct {
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"params"`
}

// PropertyGetMsg 属性查询（平台 → 设备）
type PropertyGetMsg struct {
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"params"`
}

// EventReportMsg 事件上报（设备 → 平台）
type EventReportMsg struct {
	ID    string `json:"id"`
	Event struct {
		Name   string                 `json:"name"`
		Type   string                 `json:"type"`
		Params map[string]interface{} `json:"params"`
	} `json:"event"`
}

// StatusReportMsg 状态上报（设备 → 平台）
type StatusReportMsg struct {
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"params"`
	Time   int64                  `json:"time"`
}

// ServiceCallMsg 服务调用（平台 → 设备）
type ServiceCallMsg struct {
	ID      string                 `json:"id"`
	Service string                 `json:"service"`
	Params  map[string]interface{} `json:"params"`
}

// SettingSetMsg 配置下发（平台 → 设备）
type SettingSetMsg struct {
	ID      string                 `json:"id"`
	Setting map[string]interface{} `json:"config"`
}

// SettingGetMsg 配置查询（平台 → 设备）
type SettingGetMsg struct {
	ID   string   `json:"id"`
	Keys []string `json:"keys"`
}

// DatabaseSyncMsg 数据库同步（平台 → 设备）
type DatabaseSyncMsg struct {
	ID     string             `json:"id"`
	Params DatabaseSyncParams `json:"params"`
}

// DatabaseSyncParams 数据库同步参数
type DatabaseSyncParams struct {
	Table   string                   `json:"table"`
	Records []map[string]interface{} `json:"records"`
}

// DatabaseLoadMsg 数据库读取（平台 → 设备）
type DatabaseLoadMsg struct {
	ID     string             `json:"id"`
	Params DatabaseLoadParams `json:"params"`
}

// DatabaseLoadParams 数据库读取参数
type DatabaseLoadParams struct {
	Table string `json:"table"`
}

// ResponseMessage 通用响应结构（设备 → 平台，用于各类 *\/response 应答）
type ResponseMessage struct {
	ID      string                 `json:"id"`
	Ret     int                    `json:"ret"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// ReportResponseMessage 上报应答（平台 → 设备，用于 property/event/status report 的应答）
type ReportResponseMessage struct {
	ID      string `json:"id"`
	Ret     int    `json:"ret"`
	Message string `json:"message"`
}

// OTAReportMsg OTA 升级进度上报（设备 → 平台）
type OTAReportMsg struct {
	ID       string `json:"id"`                 // 消息ID
	TaskID   string `json:"task_id"`            // 平台下发的升级任务ID
	Stage    string `json:"stage"`              // downloading / verifying / installing / rebooting / success / failed / rolled_back / cancelled
	Progress int    `json:"progress,omitempty"` // 0-100
	Error    string `json:"error,omitempty"`    // 失败原因
	Version  string `json:"version,omitempty"`  // 升级后版本号
	Time     int64  `json:"time,omitempty"`     // 毫秒时间戳
}
