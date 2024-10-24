package mqtt

import "github.com/god-jason/iot-master/boot"

func init() {
	boot.Register("mqtt", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
