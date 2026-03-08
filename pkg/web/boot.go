package web

import "github.com/busy-cloud/boat/boot"

func init() {
	boot.Register("web", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
