package broker

import (
	"os"
	"runtime"

	"github.com/busy-cloud/boat/config"
)

const MODULE = "broker"

func init() {
	config.SetDefault(MODULE, "enable", true)
	config.SetDefault(MODULE, "anonymous", false)
	config.SetDefault(MODULE, "port", 1883)

	if runtime.GOOS == "windows" {
		config.SetDefault(MODULE, "unixsock", os.TempDir()+"/boat.sock")
	} else {
		config.SetDefault(MODULE, "unixsock", "/var/run/boat.sock")
	}

	config.SetDefault(MODULE, "loglevel", "ERROR")
}
