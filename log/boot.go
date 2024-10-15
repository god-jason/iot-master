package log

import "github.com/zgwit/iot-master/boot"

func init() {
	boot.Register("log", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
