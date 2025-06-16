package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/link"
	"strings"
)

func mqttSubscribeLink() {

	mqtt.Subscribe("protocol/+/link/+/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		protocol := ss[1]
		linker := ss[3]
		link_id := ss[4]

		//查询相关的设备
		var ds []*Device
		err := db.Engine().Where("link_id=", link_id).Find(&ds)
		if err != nil {
			log.Error(err)
			return
		}

		var products []string

		var lds []*link.Device
		for _, d := range ds {
			if d.Disabled {
				continue
			}

			//记录需要挂载的设备
			lds = append(lds, &link.Device{
				Id:        d.Id,
				ProductId: d.ProductId,
				Station:   d.Station,
			})
			products = append(products, d.ProductId)

			//打开设备
			err = d.Open()
			if err != nil {
				log.Error(err)
				continue
			}
			devices.Store(d.Id, d)

			d.protocol = protocol
			d.linker = linker
		}

		//通知协议配置，TODO 这里会重复通知
		for _, pid := range products {
			cfg, err := LoadConfigure(pid, protocol)
			if err != nil {
				log.Error(err)
				continue
			}
			topic = fmt.Sprintf("protocol/%s/product/%s/config", protocol, pid)
			mqtt.Publish(topic, cfg)
		}

		//通知协议加载设备
		topic = fmt.Sprintf("protocol/%s/link/%s/%s/attach", protocol, linker, link_id)
		mqtt.Publish(topic, &lds)

	})
}
