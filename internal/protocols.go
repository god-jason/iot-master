package internal

import (
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/boat/table"
	"github.com/god-jason/iot-master/protocol"
)

var protocols lib.Map[protocol.Protocol]

func GetProtocols() []*protocol.Base {
	var b []*protocol.Base
	protocols.Range(func(name string, p *protocol.Protocol) bool {
		b = append(b, &p.Base)
		return true
	})
	return b
}

func GetRawProtocols() *lib.Map[protocol.Protocol] {
	return &protocols
}

func GetProtocol(name string) *protocol.Protocol {
	return protocols.Load(name)
}

func mqttSubscribeProtocolRegister() {
	mqtt.SubscribeStruct[protocol.Protocol](protocol.RegisterTopic, func(topic string, p *protocol.Protocol) {
		protocols.Store(p.Name, p)

		tab, err := table.Get("device")
		if err != nil {
			log.Error(err)
			return
		}

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
	})
}
