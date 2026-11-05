package types

import "time"

// Org 组织实体（代理商 / 经销商 / 门店），通过 parent_id + path 组成多级组织树
// path 为物化路径（形如 /1/3/8/），子树查询只需 path LIKE `/1/3/%`
type Org struct {
	ID          int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`
	ParentID    int64     `json:"parent_id" xorm:"parent_id bigint index default 0"` // 上级组织ID，0 表示根组织
	Path        string    `json:"path,omitempty" xorm:"path varchar(512) index"`     // 物化路径 /1/3/8/
	Name        string    `json:"name" xorm:"name varchar(128)"`                     // 组织名称
	Type        string    `json:"type,omitempty" xorm:"type varchar(32)"`            // 类型：platform / agent / dealer / store
	Level       int       `json:"level" xorm:"level int default 1"`                  // 层级，根组织为 1
	OwnerUserID string    `json:"owner_user_id,omitempty" xorm:"owner_user_id char(32) index"`
	Region      string    `json:"region,omitempty" xorm:"region varchar(128)"` // 所在区域
	Sort        int       `json:"sort,omitempty" xorm:"sort int default 0"`    // 同级排序
	Disabled    bool      `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"`
	Created     time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`
}
