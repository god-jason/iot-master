package pool

import (
	"github.com/god-jason/iot-master/config"
)

const MODULE = "pool"

func init() {
	config.Register(MODULE, "size", 10000)
}
