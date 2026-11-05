package logger

import (
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}

// Startup boot 启动入口：直接从 viper 读配置初始化 logger
// viper 没配到 output 时默认 "both"（文件 + stdout）
func Startup() error {
	level := config.GetString(MODULE, "level")
	output := config.GetString(MODULE, "output")
	logDir := config.GetString(MODULE, "log_dir")
	maxSize := config.GetInt(MODULE, "max_size")
	maxBackups := config.GetInt(MODULE, "max_backups")
	maxAge := config.GetInt(MODULE, "max_age")

	if output == "" || output == "stdout" {
		output = "both"
	}

	initWithValues(level, output, logDir, maxSize, maxBackups, maxAge)
	return nil
}

// Shutdown boot 关闭入口
func Shutdown() error { return nil }
