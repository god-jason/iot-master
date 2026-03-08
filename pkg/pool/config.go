package pool

import (
	"github.com/god-jason/iot-master/pkg/config"
)

const MODULE = "pool"

func init() {
	config.SetDefault(MODULE, "size", 10000)
}
