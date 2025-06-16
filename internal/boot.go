package internal

import (
	"github.com/busy-cloud/boat/boot"
	_ "github.com/god-jason/iot-master/device"
	_ "github.com/god-jason/iot-master/product"
	_ "github.com/god-jason/iot-master/project"
	_ "github.com/god-jason/iot-master/protocol"
	_ "github.com/god-jason/iot-master/space"
)

func init() {
	boot.Register("iot", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database", "connector"},
	})
}

func Startup() error {

	mqttSubscribeDevice()

	mqttSubscribeLink()

	return nil
}
