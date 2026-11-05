package types

import (
	"time"
)

// DeviceLog 设备日志
type DeviceLog struct {
	ID       int64       `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 日志ID
	UserID   string      `json:"user_id,omitempty" xorm:"user_id char(32) index"`     // 操作人ID → User(string)
	DeviceID string      `json:"device_id,omitempty" xorm:"device_id char(32) index"` // 设备ID → Device(string)
	Level    string      `json:"level,omitempty" xorm:"level varchar(16)"`            // 日志级别（debug/info/warn/error）
	Message  string      `json:"message,omitempty" xorm:"message varchar(512)"`       // 日志内容
	Data     interface{} `json:"data,omitempty" xorm:"data json"`                     // 附加数据（JSON）
	Created  time.Time   `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
