package maintenance

import (
	"github.com/god-jason/iot-master/pkg/boot"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Start,
		Shutdown: Stop,
		Depends:  []string{"config", "db"},
	})
}
