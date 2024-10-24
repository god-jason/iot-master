package device

import (
	"encoding/json"
	"github.com/god-jason/bucket/log"
	"github.com/god-jason/iot-master/mqtt"
	"strings"
)

type PayloadActionDown struct {
	Id         string         `json:"id"`
	Name       string         `json:"name"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type PayloadActionUp struct {
	Id     string         `json:"id"`
	Name   string         `json:"name"`
	Result string         `json:"result,omitempty"`
	Return map[string]any `json:"return,omitempty"`
}

func subscribe() {

	//属性设置
	mqtt.Subscribe("down/device/+/property", func(topic string, payload []byte) {
		var values map[string]any
		err := json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}
		ss := strings.Split(topic, "/")
		id := ss[2]
		dev := devices.Load(id)
		if dev != nil {
			err = dev.WriteMany(values)
			if err != nil {
				log.Error(err)
			}
		}

		//todo 反馈
	})

	//处理响应
	mqtt.Subscribe("down/device/+/action", func(topic string, payload []byte) {
		var down PayloadActionDown
		err := json.Unmarshal(payload, &down)
		if err != nil {
			log.Error(err)
			return
		}
		ss := strings.Split(topic, "/")
		id := ss[2]
		dev := devices.Load(id)

		var up PayloadActionUp
		up.Id = down.Id
		up.Name = down.Name

		if dev != nil {
			up.Return, err = dev.Action(down.Name, down.Parameters)
			if err != nil {
				up.Result = "ok"
			} else {
				up.Result = err.Error()
			}
		} else {
			up.Result = "设备找不到"
		}

		//上传反馈
		mqtt.Publish("up/device/"+id+"/action", &up)
	})

}
