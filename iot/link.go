package iot

import (
	"errors"
	"time"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/mqtt"
)

type Link struct {
	Id              string         `json:"id,omitempty" xorm:"pk"`
	Linker          string         `json:"linker,omitempty" xorm:"index"`
	Name            string         `json:"name,omitempty"`
	Description     string         `json:"description,omitempty"`
	Protocol        string         `json:"protocol,omitempty"`                     //通讯协议
	ProtocolOptions map[string]any `json:"protocol_options,omitempty" xorm:"json"` //通讯协议参数
	Disabled        bool           `json:"disabled,omitempty"`                     //禁用
	Online          bool           `json:"online,omitempty"`
	Error           string         `json:"error,omitempty"`
	Created         time.Time      `json:"created,omitempty,omitzero" xorm:"created"` //创建时间
}

func (l *Link) Close() error {
	token := mqtt.Publish("linker/"+l.Linker, "")
	token.Wait()
	err := token.Error()
	return err
}

var links lib.Map[Link]

func GetLink(id string) *Link {
	return links.Load(id)
}

func LoadLink(id string) (*Link, error) {
	d := &Link{}
	has, err := db.Engine().ID(id).Get(&d)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("device not exist")
	}
	links.Store(id, d)

	return d, nil
}

func UnloadLink(id string) error {
	//close?
	links.Delete(id)
	return nil
}
