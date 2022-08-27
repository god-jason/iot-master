package protocols

import (
	"fmt"
	"iot-master/connect"
	"iot-master/protocols/mitsubishi"
	"iot-master/protocols/modbus"
	"iot-master/protocols/omron"
	"iot-master/protocols/protocol"
	"iot-master/protocols/siemens"
)

var protocols = []protocol.Desc{
	modbus.DescRTU,
	modbus.DescTCP,
	omron.DescHostlink,
	omron.DescUDP,
	omron.DescTCP,
	siemens.DescS7_200_Smart,
	siemens.DescS7_200,
	siemens.DescS7_300,
	siemens.DescS7_400,
	siemens.DescS7_1200,
	siemens.DescS7_1500,
	mitsubishi.Fx_Program,
}

func Protocols() []protocol.Desc {
	return protocols
}

func Get(name string) (*protocol.Desc, error) {
	for i := 0; i < len(protocols); i++ {
		if protocols[i].Name == name {
			return &protocols[i], nil
		}
	}
	return nil, fmt.Errorf("未知协议 %s", name)
}

func Create(link connect.Tunnel, name string, options protocol.Options) (protocol.Protocol, error) {
	for _, d := range protocols {
		if d.Name == name {
			return d.Factory(link, options), nil
		}
	}
	return nil, fmt.Errorf("未知协议 %s", name)
}
