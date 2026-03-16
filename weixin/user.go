package weixin

import (
	"time"

	"github.com/god-jason/iot-master/pkg/db"
)

type User struct {
	Id        string    `json:"id" xorm:"pk"`
	OpenId    string    `json:"openid,omitempty" xorm:"'openid'"`
	UnionId   string    `json:"unionid,omitempty" xorm:"'unionid'"`
	Name      string    `json:"name,omitempty"`
	Avatar    string    `json:"avatar,omitempty"`
	Cellphone string    `json:"cellphone,omitempty"`
	Admin     bool      `json:"admin,omitempty"`
	Disabled  bool      `json:"disabled,omitempty"`
	Created   time.Time `json:"created,omitempty" xorm:"created"`
	Updated   time.Time `json:"updated,omitempty" xorm:"updated"`
}

func init() {
	db.Register(new(User))
}
