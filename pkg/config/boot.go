package config

import (
	"github.com/god-jason/iot-master/pkg/boot"
)

func init() {
	boot.Register("config", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
	})
}

func Startup() error {
	//加载配置文件
	err := Load(true)
	if err != nil {
		return err
	}

	return nil
}

func Shutdown() error {
	return nil
}
