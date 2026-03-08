package table

import (
	"github.com/god-jason/iot-master/pkg/config"
)

const MODULE = "table"

func init() {
	config.SetDefault(MODULE, "sync", true)
	config.SetDefault(MODULE, "path", "tables")
}
