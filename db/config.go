package db

import (
	"github.com/zgwit/iot-master/config"
)

const MODULE = "database"

func init() {
	config.Register(MODULE, "url", "mongodb://localhost:27017")
	config.Register(MODULE, "database", "bucket")
	config.Register(MODULE, "auth", "admin")
	config.Register(MODULE, "username", "admin")
	config.Register(MODULE, "password", "123456")
}
