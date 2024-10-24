package driver

import (
	"github.com/zgwit/iot-master/connect"
	"github.com/zgwit/iot-master/types"
)

type Factory func(conn connect.Conn, opts map[string]any) Adapter

type Driver struct {
	Name    string  `json:"name"`
	Label   string  `json:"label"`
	Factory Factory `json:"-"`

	//通道参数
	OptionForm []types.SmartField `json:"-"`

	//设备参数
	StationForm []types.SmartField `json:"-"`

	//产品参数
	ProductConfigures []ProductConfigure `json:"-"`
}

type ProductConfigure struct {
	Name  string             `json:"name"`
	Label string             `json:"label"`
	Form  []types.SmartField `json:"form"`
}
