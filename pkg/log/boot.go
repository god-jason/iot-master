package log

import "github.com/god-jason/iot-master/pkg/boot"

func init() {
	boot.Register("log", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
