package protocol

import (
	"fmt"
	"github.com/zgwit/iot-master/connect"
)

var protocols = map[string]*Protocol{}

func Protocols() []*Protocol {
	var ps []*Protocol
	for _, p := range protocols {
		ps = append(ps, p)
	}
	return ps
}

func Get(name string) (*Protocol, error) {
	if p, ok := protocols[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("协议 %s 找不到", name)
}

func Register(proto *Protocol) {
	protocols[proto.Name] = proto
}

func Create(conn connect.Conn, name string, opts map[string]any) (Adapter, error) {
	if p, ok := protocols[name]; ok {
		return p.Factory(conn, opts), nil
	}
	return nil, fmt.Errorf("协议 %s 找不到", name)
}
