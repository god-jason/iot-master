package protocol

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

var subscribes []func()

func Create(protocol *Protocol, manager MasterManager) {

	//直接注册
	Register(protocol)

	subscribes = append(subscribes, func() {

		//数据下发的回调
		writeLinkFunc := func(linker, linker_id string, data []byte) error {
			topic := "link/" + linker + "/" + linker_id + "/down"
			tkn := mqtt.Publish(topic, data)
			tkn.Wait()
			return tkn.Error()
		}

		//订阅数据
		mqtt.Subscribe("protocol/"+protocol.Name+"/link/+/+/up", func(topic string, payload []byte) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				master.OnData(payload)
			}
		})

		//连接打开，加载设备
		mqtt.Subscribe("protocol/"+protocol.Name+"/link/+/+/open", func(topic string, payload []byte) {
			ss := strings.Split(topic, "/")
			linker := ss[3]
			link_id := ss[4]
			_, err := manager.Create(linker, link_id, payload, writeLinkFunc)
			if err != nil {
				log.Error(protocol, "master create err:", err)
				return
			}
		})

		//关闭连接
		mqtt.Subscribe("protocol/"+protocol.Name+"/link/+/+/close", func(topic string, payload []byte) {
			link_id := strings.Split(topic, "/")[4]
			err := manager.Close(link_id)
			if err != nil {
				log.Error(protocol, "master close err:", err)
				return
			}
		})

		//同步
		mqtt.SubscribeStruct[SyncRequest]("protocol/"+protocol.Name+"/link/+/+/sync", func(topic string, request *SyncRequest) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				topic = "device/" + request.DeviceId + "/sync/response"
				response, err := master.OnSync(request)
				if err != nil {
					mqtt.Publish(topic, &Response{MsgId: request.MsgId, Error: err.Error()})
				} else {
					mqtt.Publish(topic, response)
				}
			}
		})

		//读
		mqtt.SubscribeStruct[ReadRequest]("protocol/"+protocol.Name+"/link/+/+/read", func(topic string, request *ReadRequest) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				topic = "device/" + request.DeviceId + "/read/response"
				response, err := master.OnRead(request)
				if err != nil {
					mqtt.Publish(topic, &Response{MsgId: request.MsgId, Error: err.Error()})
				} else {
					mqtt.Publish(topic, response)
				}
			}
		})

		//写
		mqtt.SubscribeStruct[WriteRequest]("protocol/"+protocol.Name+"/link/+/+/write", func(topic string, request *WriteRequest) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				topic = "device/" + request.DeviceId + "/write/response"
				response, err := master.OnWrite(request)
				if err != nil {
					mqtt.Publish(topic, &Response{MsgId: request.MsgId, Error: err.Error()})
				} else {
					mqtt.Publish(topic, response)
				}
			}
		})

		//操作
		mqtt.SubscribeStruct[ActionRequest]("protocol/"+protocol.Name+"/link/+/+/action", func(topic string, request *ActionRequest) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				topic = "device/" + request.DeviceId + "/action/response"
				response, err := master.OnAction(request)
				if err != nil {
					mqtt.Publish(topic, &Response{MsgId: request.MsgId, Error: err.Error()})
				} else {
					mqtt.Publish(topic, response)
				}
			}
		})

		//添加设备
		mqtt.Subscribe("protocol/"+protocol.Name+"/link/+/+/attach", func(topic string, payload []byte) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				master.OnAttach(payload)
			}
		})

		//删除设备
		mqtt.Subscribe("protocol/"+protocol.Name+"/link/+/+/detach", func(topic string, payload []byte) {
			link_id := strings.Split(topic, "/")[4]
			master := manager.Get(link_id)
			if master != nil {
				master.OnDetach(payload)
			}
		})

		//订阅产品配置
		mqtt.Subscribe("protocol/"+protocol.Name+"/product/+/config", func(topic string, payload []byte) {
			product_id := strings.Split(topic, "/")[3]
			manager.Config(product_id, payload)
		})

	})

}
