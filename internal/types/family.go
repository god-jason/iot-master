package types

import (
	"time"
)

// Family 家庭实体，用户可以创建多个家庭，用于设备分组管理
// ID 使用 int64 自增
type Family struct {
	ID      int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 家庭ID
	Name    string    `json:"name,omitempty" xorm:"name varchar(128)"`             // 家庭名称
	Icon    string    `json:"icon,omitempty" xorm:"icon varchar(512)"`             // 家庭图标
	UserID  string    `json:"user_id,omitempty" xorm:"user_id char(32) index"`     // 创建者用户ID → User(string)
	Created time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
