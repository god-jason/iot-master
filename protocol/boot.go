package protocol

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/mqtt"
)

func init() {
	boot.Register("protocol", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database"},
	})
}

func Startup() error {
	topic := "iot/protocol/register"

	//如果没有协议，则订阅，收集
	if protocols.Len() == 0 {
		mqtt.SubscribeStruct[Protocol](topic, func(topic string, protocol *Protocol) {
			protocols.Store(protocol.Name, protocol)
		})
	} else {
		protocols.Range(func(name string, p *Protocol) bool {
			mqtt.Publish("iot/protocol/register", p)
			return true
		})

		//订阅
		for _, subscribe := range subscribes {
			subscribe()
		}
	}

	return nil
}
