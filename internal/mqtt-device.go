package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/bytedance/sonic"
	"github.com/god-jason/iot-master/protocol"
	"strings"
)

func mqttSubscribeDevice() {

	mqtt.Subscribe("device/+/values", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		var values map[string]any
		err := sonic.Unmarshal(payload, &values)
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
	})

	mqtt.Subscribe("device/+/offline", func(topic string, payload []byte) {
		id := strings.Split(topic, "/")[1]
		d := devices.Load(id)
		if d != nil {
			d.Online = false
		}
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
