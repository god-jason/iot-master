package iot

import (
	"encoding/json"

	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/store"
	"github.com/god-jason/iot-master/protocol"
)

var Protocols store.FS

//var protocols lib.Map[protocol.Protocol]

func GetProtocols() ([]*protocol.Base, error) {
	var ps []*protocol.Base
	entries, err := Protocols.ReadDir("/")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			buf, err := Protocols.ReadFile(entry.Name())
			if err != nil {
				log.Error(err)
				continue
			}

			var base protocol.Base
			err = json.Unmarshal(buf, &base)
			if err != nil {
				log.Error(err)
				continue
			}
			ps = append(ps, &base)
		}
	}
	return ps, nil
}

func GetProtocol(name string) (*protocol.Protocol, error) {
	buf, err := Protocols.ReadFile(name + ".json")
	if err != nil {
		return nil, err
	}

	var p protocol.Protocol
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
