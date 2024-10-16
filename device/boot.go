package device

import "github.com/zgwit/iot-master/boot"

func init() {
	boot.Register("device", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config", "mqtt"},
	})
}

func Startup() error {

	//
	subscribe()

	return nil
}

func Shutdown() error {
	return nil
}
