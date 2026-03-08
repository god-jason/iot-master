package log

import "github.com/busy-cloud/boat/boot"

func init() {
	boot.Register("log", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
