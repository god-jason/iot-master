package history

import (
	"github.com/god-jason/iot-master/pkg/boot"
)

func init() {
	boot.Register("influxdb", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database"},
	})
}
