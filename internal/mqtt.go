package internal

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

const DeviceTopic = "device/+/"

func Startup() error {

	mqtt.Subscribe(DeviceTopic+"values", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		var values map[string]any
		err := json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}

		d := devices.Load(id)
		if d == nil {
			d = &Device{}
			has, err := db.Engine().ID(id).Get(&d.Device)
			if err != nil {
				log.Error(err)
				return
			}
			if !has {
				log.Error("device not exist")
				return
			}
			err = d.Open()
			if err != nil {
				log.Error(err)
			}

			devices.Store(id, d)
		}

		d.PutValues(values)
	})

	mqtt.Subscribe(DeviceTopic+"sync", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		var args map[string]any
		err := json.Unmarshal(payload, &args)
		if err != nil {
			log.Error(err)
			return
		}
		args["device_id"] = id

		dev := devices.Load(id)
		if dev == nil {
			return
		}

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/sync", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, args)
	})

	mqtt.Subscribe(DeviceTopic+"read", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		var args map[string]any
		err := json.Unmarshal(payload, &args)
		if err != nil {
			log.Error(err)
			return
		}
		args["device_id"] = id

		dev := devices.Load(id)
		if dev == nil {
			return
		}

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/read", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, args)
	})

	mqtt.Subscribe(DeviceTopic+"write", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		var args map[string]any
		err := json.Unmarshal(payload, &args)
		if err != nil {
			log.Error(err)
			return
		}
		args["device_id"] = id

		dev := devices.Load(id)
		if dev == nil {
			return
		}

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/write", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, args)
	})

	mqtt.Subscribe(DeviceTopic+"action", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		var args map[string]any
		err := json.Unmarshal(payload, &args)
		if err != nil {
			log.Error(err)
			return
		}
		args["device_id"] = id

		dev := devices.Load(id)
		if dev == nil {
			return
		}

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/action", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, args)
	})

	mqtt.SubscribeStruct[SyncResponse](DeviceTopic+"sync/response", func(topic string, resp *SyncResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onSyncResponse(resp)
	})

	mqtt.SubscribeStruct[ReadResponse](DeviceTopic+"read/response", func(topic string, resp *ReadResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onReadResponse(resp)
	})

	mqtt.SubscribeStruct[WriteResponse](DeviceTopic+"write/response", func(topic string, resp *WriteResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onWriteResponse(resp)
	})

	mqtt.SubscribeStruct[ActionResponse](DeviceTopic+"action/response", func(topic string, resp *ActionResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onActionResponse(resp)
	})

	return nil
}
