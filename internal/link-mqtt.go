package internal

import (
	"fmt"
	"strings"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
)

func mqttSubscribeLink() {

	mqtt.Subscribe("protocol/+/link/+/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		protocol := ss[1]
		linker := ss[3]
		link_id := ss[4]

		//查询相关的设备
		var ds []*Device
		err := db.Engine().Where("link_id=?", link_id).Find(&ds)
		if err != nil {
			log.Error(err)
			return
		}

		//无设备，则不用挂载
		if len(ds) == 0 {
			return
		}

		//var products []string
		products := make(map[string]bool)

		var lds []*LinkDevice
		for _, d := range ds {
			if d.Disabled {
				continue
			}

			//记录需要挂载的设备
			lds = append(lds, &LinkDevice{
				Id:        d.Id,
				ProductId: d.ProductId,
				Station:   d.Station,
			})
			//products = append(products, d.ProductId)
			products[d.ProductId] = true

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

		//通知协议配置
		for pid, _ := range products {
			model, err := LoadModel(pid)
			if err != nil {
				log.Error(err)
				continue
			}
			topic = fmt.Sprintf("protocol/%s/product/%s/model", protocol, pid)
			mqtt.Publish(topic, model)
		}

		//通知协议加载设备
		topic = fmt.Sprintf("protocol/%s/link/%s/%s/attach", protocol, linker, link_id)
		mqtt.Publish(topic, &lds)
	})
}
