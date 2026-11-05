// Package types 定义系统数据模型，包括 DB 表结构、API 请求/响应与运行时状态。
package types

import (
	"time"
)

// Device 设备实体
// ID 保持 string (char(32))，因为设备是对外暴露的核心实体，字符串 ID 便于 MQTT ClientID 等场景
// 归属字段（family_id / room_id / group_id / user_id）直接挂载在设备上，
// 不再需要独立的 binding 关联表（binding 表已在架构重构中移除）
type Device struct {
	ID       string `json:"id,omitempty" xorm:"id char(32) pk notnull"`              // 设备ID（char(32) 主键）
	SN       string `json:"sn,omitempty" xorm:"sn varchar(128) unique"`              // 设备序列号（全局唯一；为空时不写该列存 NULL，允许多个设备不填 SN）
	Name     string `json:"name,omitempty" xorm:"name varchar(128)"`                 // 设备名称
	Secret   string `json:"secret,omitempty" xorm:"secret varchar(128)"`             // 一机一密（MQTT 连接密码 / API Key）
	ModelID  string `json:"model_id,omitempty" xorm:"model_id char(32) index"`       // 物模型ID → Model(string)
	LinkID   string `json:"link_id,omitempty" xorm:"link_id char(32) index"`         // 所属软网关链路ID → Link(string)
	Address  string `json:"address,omitempty" xorm:"address varchar(255)"`           // 设备地址（从站号 编号 等）
	Disabled bool   `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用

	// 位置信息（可选）
	Lat     float64 `json:"lat,omitempty" xorm:"lat double"`                      // 纬度
	Lng     float64 `json:"lng,omitempty" xorm:"lng double"`                      // 经度
	GeoHash string  `json:"geo_hash,omitempty" xorm:"geo_hash varchar(16) index"` // GeoHash 索引

	// 归属字段（原 binding 表字段直接挂载到 device）
	UserID   string `json:"user_id,omitempty" xorm:"user_id char(32) index"`   // 归属用户ID → User(string)
	FamilyID int64  `json:"family_id,omitempty" xorm:"family_id bigint index"` // 所属家庭ID → Family(int64)
	RoomID   int64  `json:"room_id,omitempty" xorm:"room_id bigint index"`     // 所属房间ID → Room(int64)
	GroupID  int64  `json:"group_id,omitempty" xorm:"group_id bigint index"`   // 所属分组ID → Group(int64)

	// 固件信息（设备状态上报或 OTA 升级成功后回写）
	FirmwareVersion   string    `json:"firmware_version,omitempty" xorm:"firmware_version varchar(64)"`              // 当前固件版本
	FirmwareUpdatedAt time.Time `json:"firmware_updated_at,omitempty,zeroempty" xorm:"firmware_updated_at datetime"` // 固件版本更新时间

	Created time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
