package types

import (
	"time"
)

// Room 房间，属于家庭的子概念，用于设备按物理空间分组
// ID 使用 int64 自增
type Room struct {
	ID       int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 房间ID
	Name     string    `json:"name,omitempty" xorm:"name varchar(128)"`             // 房间名称
	FamilyID int64     `json:"family_id,omitempty" xorm:"family_id bigint index"`   // 所属家庭ID → Family(int64)
	Sort     int       `json:"sort,omitempty" xorm:"sort int default 0"`            // 排序值（越小越靠前）
	Created  time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
