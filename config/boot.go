package config

import (
	"github.com/zgwit/iot-master/boot"
)

func init() {
	boot.Register("config", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
	})
}

func Startup() error {
	//加载配置文件
	err := Load()
	if err != nil {
		//log.Error(err)
		_ = Store()
	}

	return nil
}

func Shutdown() error {
	return nil
}
