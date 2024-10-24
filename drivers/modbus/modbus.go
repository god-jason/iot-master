package modbus

import (
	"github.com/god-jason/bucket/types"
	"github.com/god-jason/iot-master/connect"
	"github.com/god-jason/iot-master/driver"
)

//var code = []types.Code{
//	{Code: 1, Label: "线圈"},
//	{Code: 2, Label: "离散输入"},
//	{Code: 3, Label: "保持寄存器"},
//	{Code: 4, Label: "输入寄存器"},
//}

var code = types.SmartField{Key: "code", Label: "功能码", Type: "select", Options: []types.SmartSelectOption{
	{Value: 1, Label: "线圈 01"},
	{Value: 2, Label: "离散输入 02"},
	{Value: 3, Label: "保持寄存器 03"},
	{Value: 4, Label: "输入寄存器 04"},
}}

var optionForm = []types.SmartField{
	{Key: "timeout", Label: "超时", Tips: "毫秒", Type: "number", Min: 1, Max: 5000, Default: 500},
	{Key: "poller_interval", Label: "轮询间隔", Tips: "秒", Type: "number", Min: 0, Max: 3600 * 24, Default: 60},
}

var pollersForm = []types.SmartField{
	code,
	{Key: "address", Label: "地址", Type: "number", Required: true, Min: 0, Max: 50000},
	{Key: "length", Label: "长度", Type: "number", Required: true, Min: 0, Max: 50000},
}

var bitPoints = []types.SmartField{
	{Key: "name", Label: "变量", Type: "text"},
	{Key: "address", Label: "地址", Type: "number", Required: true, Min: 0, Max: 50000},
}

var wordPoints = []types.SmartField{
	{Key: "name", Label: "变量", Type: "text"},
	{Key: "address", Label: "地址", Type: "number", Required: true, Min: 0, Max: 50000},
	{Key: "type", Label: "数据类型", Type: "select", Options: []types.SmartSelectOption{
		{Label: "INT16", Value: "int16"},
		{Label: "UINT16", Value: "uint16"},
		{Label: "INT32", Value: "int32"},
		{Label: "UINT32", Value: "uint32"},
		{Label: "FLOAT", Value: "float"},
		{Label: "DOUBLE", Value: "double"},
	}, Default: "uint16"},
	{Key: "be", Label: "大端", Type: "switch", Default: true},
	{Key: "rate", Label: "倍率", Type: "number", Default: 1},
	{Key: "correct", Label: "纠正", Type: "number", Default: 0},
	{Key: "bits", Label: "取位", Type: "table", Children: []types.SmartField{
		{Key: "name", Label: "变量", Type: "text", Required: true},
		{Key: "bit", Label: "位", Type: "number", Required: true, Min: 0, Max: 15},
	}},
}

var mapperForm = []types.SmartField{
	{Key: "coils", Label: "线圈 01", Type: "table", Children: bitPoints},
	{Key: "discrete_inputs", Label: "离散输入 02", Type: "table", Children: bitPoints},
	{Key: "holding_registers", Label: "保持寄存器 03", Type: "table", Children: wordPoints},
	{Key: "input_registers", Label: "输入寄存器 04", Type: "table", Children: wordPoints},
}

var stationForm = []types.SmartField{
	{Key: "slave", Label: "Modbus从站号", Type: "number", Min: 1, Max: 255, Step: 1, Default: 1},
}

var modbusRtu = &driver.Driver{
	Name:  "modbus-rtu",
	Label: "Modbus RTU",
	Factory: func(conn connect.Conn, opts map[string]any) driver.Adapter {
		return &Adapter{
			modbus:   NewRTU(conn, opts),
			devices:  make(map[string]string),
			stations: make(map[string]types.Options),
			mappers:  make(map[string]*Mapper),
			pollers:  make(map[string]*[]*Poller),
		}
	},
	OptionForm:  optionForm,
	MapperForm:  mapperForm,
	PollersForm: pollersForm,
	StationForm: stationForm,
}

var modbusTCP = &driver.Driver{
	Name:  "modbus-tcp",
	Label: "Modbus TCP",
	Factory: func(conn connect.Conn, opts map[string]any) driver.Adapter {
		return &Adapter{
			modbus:   NewTCP(conn, opts),
			devices:  make(map[string]string),
			stations: make(map[string]types.Options),
			mappers:  make(map[string]*Mapper),
			pollers:  make(map[string]*[]*Poller),
		}
	},
	OptionForm:  optionForm,
	StationForm: stationForm,
	MapperForm:  mapperForm,
	PollersForm: pollersForm,
}

type Modbus interface {
	Read(station, code uint8, addr, size uint16) ([]byte, error)
	Write(station, code uint8, addr uint16, buf []byte) error
}

func init() {
	driver.Register(modbusRtu)
	driver.Register(modbusTCP)
}
