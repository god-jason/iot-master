package apis

import "time"

// User 用户
type User struct {
	Id       string    `json:"id" xorm:"pk"`
	Name     string    `json:"name,omitempty"`
	Admin    bool      `json:"admin,omitempty"`
	TenantId string    `json:"tenant_id,omitempty"` //多租户
	Disabled bool      `json:"disabled,omitempty"`
	Created  time.Time `json:"created,omitempty" xorm:"created"`
}

// Password 密码
type Password struct {
	Id       string `json:"id" xorm:"pk"`
	Password string `json:"password"`
}

type UserLog struct {
	Id      string    `json:"id"`
	Name    string    `json:"name,omitempty"`
	Action  string    `json:"action,omitempty"`
	Client  string    `json:"client,omitempty"`
	Ip      string    `json:"ip,omitempty"`
	Created time.Time `json:"created,omitempty" xorm:"created"`
}
