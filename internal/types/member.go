package types

import (
	"time"
)

// Member 家庭成员关系，实现用户与家庭的多对多关联
// ID 使用 int64 自增
type Member struct {
	ID       int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 成员关系ID
	FamilyID int64     `json:"family_id,omitempty" xorm:"family_id bigint index"`   // 所属家庭ID → Family(int64)
	UserID   string    `json:"user_id,omitempty" xorm:"user_id char(32) index"`     // 成员用户ID → User(string)
	Admin    bool      `json:"admin,omitempty" xorm:"admin tinyint(1) default 0"`   // 是否为家庭管理员
	Created  time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
