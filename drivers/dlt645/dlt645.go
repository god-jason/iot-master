package dlt645

import (
	"github.com/god-jason/iot-master/connect"
	"github.com/god-jason/iot-master/driver"
	"github.com/god-jason/iot-master/types"
)

var stationForm = []types.SmartField{
	{Key: "slave", Label: "从站号", Type: "string", Placeholder: "12位数字"},
}

var optionForm = []types.SmartField{
	{Key: "timeout", Label: "超时", Tips: "毫秒", Type: "number", Min: 1, Max: 5000, Default: 500},
	{Key: "poller_interval", Label: "轮询间隔", Tips: "秒", Type: "number", Min: 0, Max: 3600 * 24, Default: 60},
}

var points = []types.SmartField{
	{Key: "id", Label: "标识", Type: "text"},
	{Key: "name", Label: "变量", Type: "text"},
	{Key: "label", Label: "显示", Type: "text"},
	{Key: "precision", Label: "精度", Type: "number", Required: true, Min: 0, Max: 10},
	{Key: "unit", Label: "单位", Type: "text"},
}

var mapperForm = []types.SmartField{
	{Key: "points", Label: "点位", Type: "table", Children: points},
}

type Dlt645 interface {
	Read(slave string, id string) (any, error)
}

var _dlt1997 = &driver.Driver{
	Name:  "dlt645-1997",
	Label: "DL/T645-1997",
	Factory: func(conn connect.Conn, opts map[string]any) (protocol.Adapter, error) {
		adapter := &Adapter{
			tunnel:  conn,
			dlt645:  NewDlt1997(conn, opts),
			index:   make(map[string]*Device),
			options: opts,
			mapper:  &Mapper{Points: Point1997},
		}
		err := adapter.start()
		if err != nil {
			return nil, err
		}
		return adapter, nil
	},
	OptionForm:  optionForm,
	StationForm: stationForm,
	MapperForm:  mapperForm,
}

var _dlt2007 = &protocol.Protocol{
	Name:  "dlt645-2007",
	Label: "DL/T645-2007",
	Factory: func(conn connect.Tunnel, opts map[string]any) (protocol.Adapter, error) {
		adapter := &Adapter{
			tunnel:  conn,
			dlt645:  NewDlt2007(conn, opts),
			index:   make(map[string]*Device),
			options: opts,
			mapper:  &Mapper{Points: Point2007},
		}
		err := adapter.start()
		if err != nil {
			return nil, err
		}
		return adapter, nil
	},
	OptionForm:  optionForm,
	StationForm: stationForm,
	MapperForm:  mapperForm,
}

func init() {
	protocol.Register(_dlt1997)
	protocol.Register(_dlt2007)
}
