package broker

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper 模块名
const MODULE = "broker"

func init() {
	// enable 是否启用内嵌 broker（默认开启）
	config.SetDefault(MODULE, "enable", true)
	// host 监听地址
	config.SetDefault(MODULE, "host", "0.0.0.0")
	// port 监听端口，默认 1883
	config.SetDefault(MODULE, "port", 1883)
}

// Enabled 是否启用内嵌 broker（默认 true）
func Enabled() bool { return config.GetBool(MODULE, "enable") }
