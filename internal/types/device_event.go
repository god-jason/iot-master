package types

import (
	"time"
)

// DeviceEvent 设备事件
type DeviceEvent struct {
	ID       int64       `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 事件ID
	DeviceID string      `json:"device_id,omitempty" xorm:"device_id char(32) index"` // 设备ID → Device(string)
	Type     string      `json:"type,omitempty" xorm:"type varchar(32)"`              // 事件类型
	Code     string      `json:"code,omitempty" xorm:"code varchar(64)"`              // 事件编码
	Message  string      `json:"message,omitempty" xorm:"message varchar(512)"`       // 事件消息
	Data     interface{} `json:"data,omitempty" xorm:"data json"`                     // 事件数据（JSON）
	Status   int         `json:"status,omitempty" xorm:"status int default 0"`        // 状态码（业务自定义）
	Created  time.Time   `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
