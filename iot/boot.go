package iot

import (
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/db"
	_ "github.com/god-jason/iot-master/pkg/product"
	_ "github.com/god-jason/iot-master/pkg/protocol"
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

	return addProtocolColumns()
}
