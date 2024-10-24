package server

import (
	"github.com/god-jason/bucket/lib"
	"github.com/god-jason/bucket/log"
	"github.com/god-jason/iot-master/db"
	"xorm.io/xorm"
)

var servers lib.Map[Server]

func LoadServers() error {
	var servers []*Server
	err := db.Engine.Find(&servers)
	if err != nil {
		if err == xorm.ErrNotExist {
			return nil
		}
		return err
	}
	for _, m := range servers {
		if m.Disabled {
			continue
		}
		go func(m *Server) {
			err := LoadServer(m)
			if err != nil {
				log.Error(err)
			}
		}(m)
	}
	return nil
}

func LoadServer(m *Server) error {
	servers.Store(m.Id, m)
	return m.Open()
}

func GetServer(id string) *Server {
	return servers.Load(id)
}
