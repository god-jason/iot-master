package iot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/boat/table"
	"github.com/god-jason/iot-master/protocol"
)

type Sync struct {
	Updated string `json:"updated,omitempty"`
	Created string `json:"created,omitempty"`
}

type Register struct {
	Id        string                     `json:"id,omitempty"`
	ProductId string                     `json:"product_id,omitempty"`
	Bsp       string                     `json:"bsp,omitempty"`
	Firmware  string                     `json:"firmware,omitempty"`
	Imei      string                     `json:"imei,omitempty"`
	Iccid     string                     `json:"iccid,omitempty"`
	Settings  map[string]int             `json:"settings,omitempty"`  //配置文件版本号
	Databases map[string]map[string]Sync `json:"databases,omitempty"` //数据库同步
}

func mqttSubscribeDevice() {

	//设备注册
	mqtt.SubscribeStruct[Register]("device/+/register", func(topic string, reg *Register) {
		var err error

		//查询
		d := GetDevice(reg.Id)
		if d == nil {
			d, err = LoadDevice(reg.Id)
			if err != nil {
				var dev Device
				dev.Id = reg.Id
				dev.ProductId = reg.ProductId
				dev.Online = true
				_, err = db.Engine().Insert(&dev)
				if err != nil {
					log.Error("Insert device fail", err)
					return
				}
				d, _ = LoadDevice(reg.Id)
			} else {
				d.Online = true
			}
		}

		hasSync := false

		//同步配置
		if len(reg.Settings) > 0 {
			has, err := settingSync(d.Id, reg.Settings)
			if err != nil {
				log.Error("Sync setting fail", err)
				return
			}
			if has {
				hasSync = true
			}
		}

		//同步数据库
		if len(reg.Databases) > 0 {
			has, err := databaseSync(d.Id, reg.Databases)
			if err != nil {
				log.Error("Sync database fail", err)
				return
			}
			if has {
				hasSync = true
			}
		}

		//配置和数据库更新，重启一下设备
		if hasSync {
			mqtt.Publish("device/"+d.Id+"/action/reboot", nil)
		}
	})

	mqtt.Subscribe("device/+/values", func(topic string, payload []byte) {
		var err error
		id := strings.Split(topic, "/")[1]

		d := devices.Load(id)
		if d == nil {
			d, err = LoadDevice(id)
			if err != nil {
				log.Error(err)
				return
			}
		}

		var values map[string]any
		err = json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}

		d.PutValues(values)

		//有数据就恢复上线
		if !d.Online {
			d.Online = true

			var dev Device
			dev.Online = true
			_, _ = db.Engine().ID(id).Cols("online").Update(&dev)
		}
	})

	mqtt.Subscribe("device/+/property", func(topic string, payload []byte) {
		var err error

		id := strings.Split(topic, "/")[1]

		d := devices.Load(id)
		if d == nil {
			d, err = LoadDevice(id)
			if err != nil {
				log.Error(err)
				return
			}
		}

		var props map[string]*Property
		err = json.Unmarshal(payload, &props)
		if err != nil {
			log.Error(err)
			return
		}

		//转为普通格式
		var values = make(map[string]any)
		for key, prop := range props {
			values[key] = prop.Value
		}

		d.PutValues(values)

		//有数据就恢复上线
		if !d.Online {
			d.Online = true

			var dev Device
			dev.Online = true
			_, _ = db.Engine().ID(id).Cols("online").Update(&dev)
		}
	})

	mqtt.Subscribe("device/+/online", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		d := devices.Load(id)
		if d == nil {
			_, err := LoadDevice(id)
			if err != nil {
				log.Error(err)
				return
			}
		} else {
			d.Online = true
		}

		var dev Device
		dev.Online = true
		_, _ = db.Engine().ID(id).Cols("online").Update(&dev)
	})

	mqtt.Subscribe("device/+/offline", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		d := devices.Load(id)
		if d != nil {
			d.Online = false
		}

		var dev Device
		dev.Online = false
		_, _ = db.Engine().ID(id).Cols("online").Update(&dev)
		_, _ = db.Engine().Where("gateway_id=?", id).Cols("online").Update(&dev) //子设备也掉线
	})

	//监听总线消息，客户端断开，则视为下线
	mqtt.Subscribe("client/+/disconnect", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		d := devices.Load(id)
		if d != nil {
			mqtt.Publish("device/"+id+"/offline", nil)
		}
	})

	mqtt.Subscribe("device/+/log", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]

		tab, err := table.Get("device_log")
		if err != nil {
			return
		}

		_, _ = tab.Insert(map[string]interface{}{
			"device_id": id,
			"content":   string(payload),
		})
	})

	mqtt.Subscribe("device/+/log/+", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		user_id := strings.Split(topic, "/")[3]

		tab, err := table.Get("device_log")
		if err != nil {
			return
		}

		_, _ = tab.Insert(map[string]interface{}{
			"user_id":   user_id,
			"device_id": id,
			"content":   string(payload),
		})
	})

	mqtt.SubscribeStruct[protocol.SyncRequest]("device/+/sync", func(topic string, request *protocol.SyncRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/link/%s/%s/sync", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.ReadRequest]("device/+/read", func(topic string, request *protocol.ReadRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/link/%s/%s/read", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.WriteRequest]("device/+/write", func(topic string, request *protocol.WriteRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/link/%s/%s/write", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.ActionRequest]("device/+/action", func(topic string, request *protocol.ActionRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/link/%s/%s/action", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.SyncResponse]("device/+/sync/response", func(topic string, resp *protocol.SyncResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onSyncResponse(resp)
	})

	mqtt.SubscribeStruct[protocol.ReadResponse]("device/+/read/response", func(topic string, resp *protocol.ReadResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onReadResponse(resp)
	})

	mqtt.SubscribeStruct[protocol.WriteResponse]("device/+/write/response", func(topic string, resp *protocol.WriteResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onWriteResponse(resp)
	})

	mqtt.SubscribeStruct[protocol.ActionResponse]("device/+/action/response", func(topic string, resp *protocol.ActionResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onActionResponse(resp)
	})
}
