package weixin

import (
	"github.com/god-jason/iot-master/pkg/config"
)

const MODULE = "weixin"

func init() {
	config.SetDefault(MODULE, "appid", "")
	config.SetDefault(MODULE, "secret", "")
}
