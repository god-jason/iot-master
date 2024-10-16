package admin

import "github.com/zgwit/iot-master/config"

const MODULE = "admin"

func init() {
	config.Register(MODULE, "password", md5hash("123456"))
}
