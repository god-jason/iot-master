package protocol

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

func Create(protocol *Protocol, manager MasterManager) {

	writeLinkFunc := func(linker, linker_id string, data []byte) error {
		topic := "link/" + linker + "/" + linker_id + "/down"
		tkn := mqtt.Publish(topic, data)
		tkn.Wait()
		return tkn.Error()
	}

	//订阅数据
	mqtt.Subscribe(protocol.Name+"/+/+/up", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		link_id := ss[2]
		master := manager.Get(link_id)
		if master != nil {
			master.OnData(payload)
		}
	})

	//连接打开，加载设备
	mqtt.Subscribe(protocol.Name+"/+/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		linker := ss[1]
		link_id := ss[2]
		_, err := manager.Create(linker, link_id, payload, writeLinkFunc)
		if err != nil {
			log.Error(protocol, "master create err:", err)
			return
		}
	})

	//关闭连接
	mqtt.Subscribe(protocol.Name+"/+/+/close", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		link_id := ss[2]
		err := manager.Close(link_id)
		if err != nil {
			log.Error(protocol, "master close err:", err)
			return
		}
	})

}
