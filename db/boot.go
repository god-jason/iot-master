package db

import "github.com/god-jason/iot-master/boot"

func init() {
	boot.Register("database", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
