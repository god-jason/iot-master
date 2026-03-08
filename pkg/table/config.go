package table

import (
	"github.com/busy-cloud/boat/config"
)

const MODULE = "table"

func init() {
	config.SetDefault(MODULE, "sync", true)
	config.SetDefault(MODULE, "paths", []string{})
}
