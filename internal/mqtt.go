package internal

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/protocol"
	"strings"
)

const DeviceTopic = "device/+/"

func Startup() error {

	mqtt.Subscribe(DeviceTopic+"values", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		var values map[string]any
		err := json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}

		d := devices.Load(id)
		if d == nil {
			d, err = LoadDevice(id)
			if err != nil {
				log.Error(err)
				return
			}
		}

		d.PutValues(values)
	})

	mqtt.Subscribe(DeviceTopic+"online", func(topic string, payload []byte) {
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
	})

	mqtt.Subscribe(DeviceTopic+"offline", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		d := devices.Load(id)
		if d != nil {
			d.Online = false
		}
	})

	mqtt.SubscribeStruct[protocol.SyncRequest](DeviceTopic+"sync", func(topic string, request *protocol.SyncRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/sync", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.ReadRequest](DeviceTopic+"read", func(topic string, request *protocol.ReadRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/read", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.WriteRequest](DeviceTopic+"write", func(topic string, request *protocol.WriteRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/write", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.ActionRequest](DeviceTopic+"action", func(topic string, request *protocol.ActionRequest) {
		ss := strings.Split(topic, "/")
		id := ss[1]

		dev := devices.Load(id)
		if dev == nil {
			mqtt.Publish(topic+"/response", &protocol.Response{MsgId: request.MsgId, Error: "设备未上线"})
			return
		}

		request.DeviceId = id

		//转发给具体协议
		topic = fmt.Sprintf("protocol/%s/%s/%s/action", dev.protocol, dev.linker, dev.LinkId)
		mqtt.Publish(topic, request)
	})

	mqtt.SubscribeStruct[protocol.SyncResponse](DeviceTopic+"sync/response", func(topic string, resp *protocol.SyncResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onSyncResponse(resp)
	})

	mqtt.SubscribeStruct[protocol.ReadResponse](DeviceTopic+"read/response", func(topic string, resp *protocol.ReadResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onReadResponse(resp)
	})

	mqtt.SubscribeStruct[protocol.WriteResponse](DeviceTopic+"write/response", func(topic string, resp *protocol.WriteResponse) {
		ss := strings.Split(topic, "/")
		id := ss[1]
		dev := devices.Load(id)
		if dev == nil {
			return
		}
		dev.onWriteResponse(resp)
	})

	mqtt.SubscribeStruct[protocol.ActionResponse](DeviceTopic+"action/response", func(topic string, resp *protocol.ActionResponse) {
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
