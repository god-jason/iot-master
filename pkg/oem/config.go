package oem

import (
	"github.com/god-jason/iot-master/pkg/config"
)

const MODULE = "oem"

func init() {
	config.SetDefault(MODULE, "name", "BOAT")
	config.SetDefault(MODULE, "logo", "")
}
