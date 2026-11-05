package types

import "time"

// DeviceOrg 设备与组织的多对多绑定关系
// 一台设备可同时绑定多个组织；组织树上的分配本质是新增一条绑定记录
type DeviceOrg struct {
	ID       int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`
	DeviceID string    `json:"device_id" xorm:"device_id char(32) index"`
	OrgID    int64     `json:"org_id" xorm:"org_id bigint index"`
	Created  time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`
}
