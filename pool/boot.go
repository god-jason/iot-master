package pool

import "github.com/zgwit/iot-master/boot"

func init() {
	boot.Register("pool", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
