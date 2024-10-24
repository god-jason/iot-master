package web

import "github.com/god-jason/iot-master/boot"

func init() {
	boot.Register("web", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
