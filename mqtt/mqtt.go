package mqtt

import (
	"encoding/json"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/god-jason/bucket/config"
	"github.com/god-jason/bucket/log"
	"github.com/god-jason/bucket/pool"
)

var Client paho.Client

func Startup() error {

	//正常流程
	opts := paho.NewClientOptions()
	opts.AddBroker(config.GetString(MODULE, "url"))
	opts.SetClientID(config.GetString(MODULE, "clientId"))
	opts.SetUsername(config.GetString(MODULE, "username"))
	opts.SetPassword(config.GetString(MODULE, "password"))
	opts.SetConnectRetry(true) //重试

	opts.SetKeepAlive(20)

	//重连时，恢复订阅
	opts.SetCleanSession(false)
	opts.SetResumeSubs(true)

	//加上订阅处理
	opts.SetOnConnectHandler(func(client paho.Client) {
		//for topic, _ := range subs {
		//	Client.Subscribe(topic, 0, func(client paho.Client, message paho.Message) {
		//
		//		go func() {
		//			//依次处理回调
		//			if cbs, ok := subs[topic]; ok {
		//				for _, cb := range cbs {
		//					cb(message.Topic(), message.Payload())
		//				}
		//			}
		//		}()
		//	})
		//}
	})

	Client = paho.NewClient(opts)
	token := Client.Connect()
	//token.Wait()
	return token.Error()
}

func Shutdown() error {
	Client.Disconnect(0)
	return nil
}

func Publish(topic string, payload any) paho.Token {
	bytes, _ := json.Marshal(payload)
	return Client.Publish(topic, 0, false, bytes)
}

func Subscribe(filter string, cb func(topic string, payload []byte)) paho.Token {
	return Client.Subscribe(filter, 0, func(client paho.Client, message paho.Message) {
		err := pool.Insert(func() {
			//c(message.Topic(), &value)
			cb(message.Topic(), message.Payload())
		})
		if err != nil {
			log.Error(err)
			return
		}
	})
}

func SubscribeStruct[T any](filter string, cb func(topic string, data *T)) paho.Token {
	return Client.Subscribe(filter, 0, func(client paho.Client, message paho.Message) {
		err := pool.Insert(func() {
			var value T
			if len(message.Payload()) > 0 {
				err := json.Unmarshal(message.Payload(), &value)
				if err != nil {
					log.Error(err)
					return
				}
			}
			cb(message.Topic(), &value)
		})
		if err != nil {
			log.Error(err)
			return
		}
	})
}

var subs = map[string]any{}

func SubscribeExt[T any](filter string, cb func(topic string, value *T)) {

	var cbs []func(topic string, value *T)

	//重复订阅，直接入列
	if callbacks, ok := subs[filter]; ok {
		cbs = callbacks.([]func(topic string, value *T))
		subs[filter] = append(cbs, cb)
		return
	}

	subs[filter] = append(cbs, cb)

	//初次订阅
	Client.Subscribe(filter, 0, func(client paho.Client, message paho.Message) {
		cbs := subs[filter]
		cs := cbs.([]func(topic string, value *T))

		//解析JSON
		var value T
		if len(message.Payload()) > 0 {
			err := json.Unmarshal(message.Payload(), &value)
			if err != nil {
				log.Error(err)
				return
			}
		}

		//回调
		for _, c := range cs {
			if pool.Pool == nil {
				go c(message.Topic(), &value)
				continue
			}
			//放入线程池处理
			err := pool.Insert(func() {
				c(message.Topic(), &value)
			})
			if err != nil {
				log.Error(err)
				go c(message.Topic(), &value)
			}
		}
	})
}
