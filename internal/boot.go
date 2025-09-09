package internal

import (
	"github.com/busy-cloud/boat/boot"
	_ "github.com/god-jason/iot-master/product"
	_ "github.com/god-jason/iot-master/protocol"
)

func init() {
	boot.Register("iot", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database", "protocol"},
	})
}

func Startup() error {

	mqttSubscribeDevice()

	mqttSubscribeLink()

	mqttSubscribeProtocolRegister()

	return nil
}
