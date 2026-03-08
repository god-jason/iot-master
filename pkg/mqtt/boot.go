package mqtt

import "github.com/busy-cloud/boat/boot"

func init() {
	boot.Register("mqtt", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config", "broker", "pool"},
	})
}
