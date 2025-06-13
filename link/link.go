package link

import (
	"github.com/busy-cloud/boat/db"
	"time"
)

func init() {
	db.Register(&Link{})
}

type Link struct {
	Id     string `json:"id,omitempty" xorm:"pk"`
	Linker string `json:"linker,omitempty" xorm:"index"`
	//ServerId        string         `json:"server_id,omitempty" xorm:"index"`
	Name            string         `json:"name,omitempty"`
	Description     string         `json:"description,omitempty"`
	Protocol        string         `json:"protocol,omitempty"`                        //通讯协议
	ProtocolOptions map[string]any `json:"protocol_options,omitempty" xorm:"json"`    //通讯协议参数
	Disabled        bool           `json:"disabled,omitempty"`                        //禁用
	Created         time.Time      `json:"created,omitempty,omitzero" xorm:"created"` //创建时间
}

type Status struct {
	Running bool   `json:"running,omitempty" xorm:"-"`
	Error   string `json:"error,omitempty" xorm:"-"`
}
