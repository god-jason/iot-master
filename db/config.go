package db

import (
	"github.com/zgwit/iot-master/config"
	"github.com/zgwit/iot-master/lib"
)

const MODULE = "database"

func init() {
	config.Register(MODULE, "type", "sqlite3")
	config.Register(MODULE, "url", lib.AppName()+".db") //"root:root@tcp(localhost:3306)/master?charset=utf8"
	config.Register(MODULE, "debug", false)
	config.Register(MODULE, "sync", true)
}
