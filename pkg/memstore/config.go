package memstore

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper 模块名（yaml 里叫 "redis"，历史原因）
const MODULE = "redis"

func init() {
	config.SetDefault(MODULE, "enable", false)
	config.SetDefault(MODULE, "host", "127.0.0.1")
	config.SetDefault(MODULE, "port", 6379)
	config.SetDefault(MODULE, "password", "")
	config.SetDefault(MODULE, "db", 0)
}

// Enabled 是否启用 Redis（默认 false，走内存）
func Enabled() bool { return config.GetBool(MODULE, "enable") }
