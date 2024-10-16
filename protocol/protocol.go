package protocol

import (
	"github.com/god-jason/bucket/types"
	"github.com/zgwit/iot-master/connect"
)

type Factory func(conn connect.Conn, opts map[string]any) Adapter

type Protocol struct {
	Name    string  `json:"name"`
	Label   string  `json:"label"`
	Factory Factory `json:"-"`

	//通道参数
	OptionForm []types.SmartField `json:"-"`

	//设备参数
	StationForm []types.SmartField `json:"-"`

	//轮询器
	PollersForm []types.SmartField `json:"-"`

	//映射表
	MapperForm []types.SmartField `json:"-"`
}
