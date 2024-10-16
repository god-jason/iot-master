package cjt188

import (
	"github.com/zgwit/iot-master/connect/tunnel"
	"github.com/zgwit/iot-master/driver"
	"github.com/zgwit/iot-master/types"
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

type cjt188 interface {
	Read(slave string, id string) (any, error)
}

var _cjt2004 = &driver.Driver{
	Name:  "cjt188-2004",
	Label: "CJ/T188-2004",
	Factory: func(conn tunnel.Tunnel, opts map[string]any) (driver.Adapter, error) {
		adapter := &Adapter{
			tunnel:  conn,
			cjt188:  NewCjt2004(conn, opts),
			index:   make(map[string]*Device),
			options: opts,
			mapper:  &Mapper{Points: Point2004},
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

var _cjt2018 = &driver.Driver{
	Name:  "cjt188-2018",
	Label: "CJ/T188-2018",
	Factory: func(conn tunnel.Tunnel, opts map[string]any) (driver.Adapter, error) {
		adapter := &Adapter{
			tunnel:  conn,
			cjt188:  NewCjt2018(conn, opts),
			index:   make(map[string]*Device),
			options: opts,
			mapper:  &Mapper{Points: Point2018},
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
	driver.Register(_cjt2004)
	driver.Register(_cjt2018)
}
