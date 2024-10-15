package db

import "github.com/zgwit/iot-master/boot"

func init() {
	boot.Register("database", &boot.Task{
		Startup:  Open,
		Shutdown: Close,
		Depends:  []string{"config", "log"},
	})
}
