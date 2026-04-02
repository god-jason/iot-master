package iot

import (
	"encoding/json"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/log"
	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/smart"
	"github.com/god-jason/iot-master/pkg/store"
	"github.com/god-jason/iot-master/pkg/table"
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

				//设备表
				sql := db.Engine().Dialect().AddColumnSQL("device", col)
				_, _ = db.Engine().Exec(sql)

				//内联设备表
				sql = db.Engine().Dialect().AddColumnSQL("inline", col)
				_, _ = db.Engine().Exec(sql)
			}
		}
	}

	return nil
}
