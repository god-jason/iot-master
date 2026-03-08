package pool

import "github.com/busy-cloud/boat/boot"

func init() {
	boot.Register("pool", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
