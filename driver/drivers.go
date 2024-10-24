package driver

import (
	"fmt"
	"github.com/god-jason/iot-master/connect"
)

var drivers = map[string]*Driver{}

func Drivers() []*Driver {
	var ps []*Driver
	for _, p := range drivers {
		ps = append(ps, p)
	}
	return ps
}

func Get(name string) (*Driver, error) {
	if p, ok := drivers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("协议 %s 找不到", name)
}

func Register(proto *Driver) {
	drivers[proto.Name] = proto
}

func Create(conn connect.Conn, name string, opts map[string]any) (Adapter, error) {
	if p, ok := drivers[name]; ok {
		return p.Factory(conn, opts), nil
	}
	return nil, fmt.Errorf("协议 %s 找不到", name)
}
