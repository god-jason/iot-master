package internal

import (
	"fmt"
	"strings"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/spf13/cast"
)

func mqttSubscribeLink() {

	mqtt.Subscribe("protocol/+/link/+/+/open", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		protocol := ss[1]
		linker := ss[3]
		link_id := ss[4]

		//查询相关的设备
		var ds []map[string]any
		err := db.Engine().Table("device").Where("link_id=?", link_id).Find(&ds)
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

		var lds []map[string]any
		for _, d := range ds {

			var dev Device
			dev.Id = cast.ToString(d["id"])
			dev.Name = cast.ToString(d["name"])
			dev.ProductId = cast.ToString(d["product_id"])
			dev.LinkId = cast.ToString(d["link_id"])
			dev.Disabled = cast.ToBool(d["disabled"]) //这里只有手动转。。。

			//err := mapstructure.WeakDecode(d, &dev)
			//dev, err := map2struct[Device](d)

			//过滤禁用的
			if dev.Disabled {
				continue
			}

			//记录需要挂载的设备
			lds = append(lds, d)
			//products = append(products, d.ProductId)
			products[dev.ProductId] = true

			//打开设备

			err = dev.Open()
			if err != nil {
				log.Error(err)
				continue
			}
			devices.Store(dev.Id, &dev)

			dev.protocol = protocol
			dev.linker = linker
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
