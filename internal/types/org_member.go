package types

import "time"

// OrgMember 组织成员（负责人之外可共同管理该组织设备与子级组织的成员）
// 与 org.owner_user_id 一起构成组织管理团队
type OrgMember struct {
	ID      int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`           // 主键
	OrgID   int64     `json:"org_id" xorm:"org_id bigint index"`                   // 组织ID
	UserID  string    `json:"user_id" xorm:"user_id char(32) index"`               // 成员用户ID
	Created time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 加入时间
}
