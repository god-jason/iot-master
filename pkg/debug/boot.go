package debug

import "github.com/god-jason/iot-master/pkg/boot"

func init() {
	boot.Register("debug", &boot.Task{
		Startup: Startup,
		Depends: []string{"config", "web"},
	})
}
