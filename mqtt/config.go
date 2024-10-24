package mqtt

import (
	"github.com/god-jason/bucket/config"
	"github.com/god-jason/bucket/lib"
)

const MODULE = "mqtt"

func init() {
	config.Register(MODULE, "url", "mqtt://localhost:1843")
	config.Register(MODULE, "clientId", lib.RandomString(12))
	config.Register(MODULE, "username", "")
	config.Register(MODULE, "password", "")
}
