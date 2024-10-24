package mqtt

import "github.com/god-jason/bucket/boot"

func init() {
	boot.Register("mqtt", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}
