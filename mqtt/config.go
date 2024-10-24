package mqtt

import (
	"github.com/god-jason/iot-master/config"
	"github.com/god-jason/iot-master/lib"
)

const MODULE = "mqtt"

func init() {
	config.Register(MODULE, "url", "mqtt://localhost:1843")
	config.Register(MODULE, "clientId", lib.RandomString(12))
	config.Register(MODULE, "username", "")
	config.Register(MODULE, "password", "")
}
