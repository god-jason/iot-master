package types

import (
	"time"
)

// Group 设备分组，支持树形结构（通过ParentID实现）
// ID 使用 int64 自增
type Group struct {
	ID       int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 分组ID
	Name     string    `json:"name,omitempty" xorm:"name varchar(128)"`             // 分组名称
	ParentID int64     `json:"parent_id,omitempty" xorm:"parent_id bigint index"`   // 上级分组ID → Group(int64), 0 表示根
	Created  time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
