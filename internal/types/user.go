package types

import (
	"time"
)

// User 用户实体，存储平台用户的基础信息
type User struct {
	ID        string    `json:"id,omitempty" xorm:"id char(32) pk notnull"`              // 用户ID
	Username  string    `json:"username,omitempty" xorm:"username varchar(128) unique"`  // 用户名（唯一）
	Password  string    `json:"password,omitempty" xorm:"password varchar(255)"`         // 密码（bcrypt 哈希）
	Nickname  string    `json:"nickname,omitempty" xorm:"nickname varchar(128)"`         // 昵称
	Avatar    string    `json:"avatar,omitempty" xorm:"avatar varchar(512)"`             // 头像地址
	Email     string    `json:"email,omitempty" xorm:"email varchar(255)"`               // 邮箱
	Cellphone string    `json:"cellphone,omitempty" xorm:"cellphone varchar(32)"`        // 手机号
	Disabled  bool      `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Lock      bool      `json:"lock,omitempty" xorm:"lock tinyint(1) default 0"`         // 是否锁定（如密码错误过多）
	Admin     bool      `json:"admin,omitempty" xorm:"admin tinyint(1) default 0"`       // 是否平台管理员
	Platform  bool      `json:"platform,omitempty" xorm:"platform tinyint(1) default 0"` // 是否平台账号（可登录管理后台，普通用户仅限移动端）
	OpenID    string    `json:"openid,omitempty" xorm:"openid varchar(128)"`             // 微信小程序 OpenID
	UnionID   string    `json:"unionid,omitempty" xorm:"unionid varchar(128)"`           // 微信开放平台 UnionID
	AppOpenID string    `json:"app_openid,omitempty" xorm:"app_openid varchar(128)"`     // 移动应用 OpenID
	Balance   int64     `json:"balance,omitempty" xorm:"balance bigint default 0"`       // 账户余额
	Created   time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`     // 创建时间
	Updated   time.Time `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`     // 更新时间
	ParentID  string    `json:"parent_id,omitempty" xorm:"parent_id char(32) index"`     // 上级用户ID（推荐人/父账号）
}
