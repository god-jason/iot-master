package logger

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper 模块名
const MODULE = "log"

func init() {
	config.SetDefault(MODULE, "level", "info")
	config.SetDefault(MODULE, "output", "stdout")
	config.SetDefault(MODULE, "log_dir", "./logs")
	config.SetDefault(MODULE, "max_size", 100)
	config.SetDefault(MODULE, "max_backups", 10)
	config.SetDefault(MODULE, "max_age", 30)
}
