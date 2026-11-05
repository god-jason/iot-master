package types

import (
	"time"
)

// 场景运行模式常量
const (
	SceneModeOnline  = "online"  // 在线模式：由 master 服务器执行
	SceneModeOffline = "offline" // 离线模式：由本地 MQTT 网关执行
)

// 场景动作类型常量
const (
	SceneActionProperty = "property" // 下发设备属性
	SceneActionService  = "service"  // 调用设备服务
)

// Scene 智能家居场景
// ID 使用 int64 自增；FamilyID→Family(int64)
type Scene struct {
	ID       int64         `json:"id,omitempty" xorm:"id bigint pk autoincr"`               // 场景ID
	FamilyID int64         `json:"family_id,omitempty" xorm:"family_id bigint index"`       // 所属家庭ID → Family(int64)
	Name     string        `json:"name,omitempty" xorm:"name varchar(128)"`                 // 场景名称
	Icon     string        `json:"icon,omitempty" xorm:"icon varchar(512)"`                 // 场景图标
	Desc     string        `json:"desc,omitempty" xorm:"desc varchar(512)"`                 // 场景描述
	Mode     string        `json:"mode,omitempty" xorm:"mode varchar(16)"`                  // 运行模式（online/offline）
	Disabled bool          `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Actions  []SceneAction `json:"actions,omitempty" xorm:"actions mediumtext json"`        // 场景动作列表（JSON）
	Created  time.Time     `json:"created,omitempty,zeroempty" xorm:"created datetime"`     // 创建时间
	Updated  time.Time     `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`     // 更新时间
}

// SceneAction 场景中的单个设备动作（内嵌为 JSON，不需要 xorm tag）
type SceneAction struct {
	DeviceID   string                 `json:"device_id,omitempty"`  // → Device(string)
	Type       string                 `json:"type,omitempty"`       // 动作类型（property / service）
	Properties map[string]interface{} `json:"properties,omitempty"` // 下发的属性键值对（Type=property 时使用）
	Service    string                 `json:"service,omitempty"`    // 服务名称（Type=service 时使用）
	Params     map[string]interface{} `json:"params,omitempty"`     // 服务入参（Type=service 时使用）
}

// SceneActionResult 单个场景动作的执行结果（只用于函数返回值，不落库）
type SceneActionResult struct {
	DeviceID string `json:"device_id,omitempty"` // 目标设备ID
	Type     string `json:"type,omitempty"`      // 动作类型
	Success  bool   `json:"success"`             // 是否执行成功
	Message  string `json:"message,omitempty"`   // 结果消息或失败原因
}

// CallResult 设备指令的执行结果（与下行报文中的 ret/message 对应）
type CallResult struct {
	Ret     int    // 返回码：1 成功，0 失败
	Message string // 设备返回的消息或失败原因
}

// DeviceCaller 设备侧能力接口（iot 规则引擎、scene 场景引擎共用）
type DeviceCaller interface {
	// IsOnline 判断设备是否在线
	IsOnline(deviceID string) bool
	// SetProperty 下发设备属性
	SetProperty(deviceID string, params map[string]interface{}) (*CallResult, error)
	// Control 调用设备服务
	Control(deviceID, service string, params map[string]interface{}) (*CallResult, error)
}
