package oem

import (
	"github.com/busy-cloud/boat/config"
)

const MODULE = "oem"

func init() {
	config.SetDefault(MODULE, "name", "BOAT")
	config.SetDefault(MODULE, "logo", "")
}
