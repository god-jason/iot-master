package internal

import (
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
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

		//向数据库中添加字段
		for _, field := range p.DeviceExtendColumns {
			col := field.ToColumn()
			sql := db.Engine().Dialect().AddColumnSQL("device", col)
			_, err := db.Engine().Exec(sql)
			if err != nil {
				log.Error(err)
			}
		}
	})
}
