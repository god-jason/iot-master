package service

import (
	"github.com/god-jason/iot-master/config"
	"github.com/god-jason/iot-master/lib"
)

const MODULE = "service"

func init() {
	config.Register(MODULE, "name", lib.AppName())
	config.Register(MODULE, "display", "物联大师")
	config.Register(MODULE, "description", "物联大师服务")
	config.Register(MODULE, "arguments", []string{})
}
