package pool

import (
	"github.com/busy-cloud/boat/config"
)

const MODULE = "pool"

func init() {
	config.SetDefault(MODULE, "size", 10000)
}
