package admin

import "github.com/god-jason/bucket/config"

const MODULE = "admin"

func init() {
	config.Register(MODULE, "password", md5hash("123456"))
}
