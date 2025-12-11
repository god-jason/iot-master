package iot

import (
	"encoding/json"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/smart"
	"github.com/busy-cloud/boat/store"
	"github.com/busy-cloud/boat/table"
	"github.com/god-jason/iot-master/protocol"
)

var Protocols store.FS

//var protocols lib.Map[protocol.Protocol]

var _deviceExtendFields []*smart.Field

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

func addProtocolColumns() error {
	tab, err := table.Get("device")
	if err != nil {
		return err
	}

	entries, err := Protocols.ReadDir("/")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			buf, err := Protocols.ReadFile(entry.Name())
			if err != nil {
				return err
			}

			var p protocol.Protocol
			err = json.Unmarshal(buf, &p)
			if err != nil {
				return err
			}

			//扩展列
			for _, field := range p.DeviceExtendFields {
				_deviceExtendFields = append(_deviceExtendFields, field)
			}

			//数据库扩展
			for _, field := range p.DeviceExtendColumns {
				tab.AddColumn(field) //添加到字段定义中

				//向数据库表定义中添加字段 TODO 存在冗余添加了
				col := field.ToColumn()
				sql := db.Engine().Dialect().AddColumnSQL("device", col)
				_, err := db.Engine().Exec(sql)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}

	return nil
}
