package db

import (
	"github.com/busy-cloud/boat/config"
)

const MODULE = "database"

func init() {
	config.SetDefault(MODULE, "type", "mysql")
	config.SetDefault(MODULE, "url", "root:123456@tcp(localhost:3306)/boat?charset=utf8")
	config.SetDefault(MODULE, "debug", false)
	config.SetDefault(MODULE, "sync", true)
}

func SetDatabaseUrl(url string) {
	config.Set(MODULE, "url", url)
}
