package iot

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/db"
	_ "github.com/god-jason/iot-master/product"
	_ "github.com/god-jason/iot-master/protocol"
)

func init() {
	boot.Register("iot", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database", "table"},
	})
}

func Startup() error {

	//开机全部下线，等待逐一上线
	var dev Device
	dev.Online = false
	_, _ = db.Engine().Where("online=1").Cols("online").Update(&dev)

	mqttSubscribeDevice()

	mqttSubscribeLink()

	mqttSubscribeProtocolRegister()

	return nil
}
