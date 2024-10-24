package pool

import "github.com/god-jason/iot-master/boot"

func init() {
	boot.Register("pool", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
