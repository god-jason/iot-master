package product

import (
	"github.com/zgwit/iot-master/base"
	"github.com/zgwit/iot-master/db"
	"time"
)

func init() {
	db.Register(new(Product))
}

type Product struct {
	Id          string    `json:"id" xorm:"pk"`          //ID
	Icon        string    `json:"icon,omitempty"`        //图标
	Name        string    `json:"name,omitempty"`        //名称
	Url         string    `json:"url,omitempty"`         //链接
	Protocol    string    `json:"driver,omitempty"`      //协议
	Description string    `json:"description,omitempty"` //说明
	Keywords    []string  `json:"keywords,omitempty"`    //关键字
	Created     time.Time `json:"created" xorm:"created"`

	operators []*base.Action
}

func (p *Product) Init() error {
	for _, a := range p.operators {
		err := a.Init()
		if err != nil {
			return err
		}
	}
	return nil
}
