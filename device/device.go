package device

import (
	"github.com/god-jason/iot-master/product"
	"time"
)

type Device struct {
	Id          string         `json:"id,omitempty" xorm:"pk"`
	ProductId   string         `json:"product_id,omitempty" xorm:"index"`
	LinkId      string         `json:"link_id,omitempty" xorm:"index"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Station     map[string]any `json:"station,omitempty" xorm:"json"` //从站信息（协议定义表单）
	Disabled    bool           `json:"disabled,omitempty"`            //禁用
	Created     time.Time      `json:"created,omitempty" xorm:"created"`
}

type DeviceModel struct {
	Id         string               `json:"id,omitempty" xorm:"pk"`
	Validators []*product.Validator `json:"validators,omitempty" xorm:"json"`
	Created    time.Time            `json:"created,omitempty" xorm:"created"`
}

type Status struct {
	Online bool   `json:"online,omitempty"`
	Error  string `json:"error,omitempty"`
}
