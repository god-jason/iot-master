package web

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper 模块名
const MODULE = "web"

func init() {
	config.SetDefault(MODULE, "host", "0.0.0.0")
	config.SetDefault(MODULE, "port", 8080)
	config.SetDefault(MODULE, "static", "")
	config.SetDefault(MODULE, "https", false)
	config.SetDefault(MODULE, "cert_file", "")
	config.SetDefault(MODULE, "key_file", "")
	config.SetDefault(MODULE, "lets_encrypt", false)
	config.SetDefault(MODULE, "domains", []string{})
	config.SetDefault(MODULE, "email", "")
	config.SetDefault(MODULE, "debug", false)
}
