package app

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/web"
	_ "github.com/god-jason/iot-master/device"
	_ "github.com/god-jason/iot-master/product"
	_ "github.com/god-jason/iot-master/project"
	_ "github.com/god-jason/iot-master/protocol"
	_ "github.com/god-jason/iot-master/space"
)

func init() {
	boot.Register("app", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"log", "web", "database"},
	})
}

func Startup() error {

	err := LoadAll()
	if err != nil {
		return err
	}

	err = LoadLicenses()
	if err != nil {
		return err
	}

	//注册APP代理
	web.Engine().Use(Proxy)

	return nil
}

func Shutdown() error {
	apps.Range(func(name string, item *App) bool {
		if item.zipReader != nil {
			_ = item.zipReader.Close()
		}
		return true
	})
	return nil
}
