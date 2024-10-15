package db

import "github.com/zgwit/iot-master/boot"

func init() {
	boot.Register("database", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
