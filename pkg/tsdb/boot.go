package tsdb

import (
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/logger"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}

func Startup() error {
	drv := Driver()
	if !Enabled() {
		drv = "memory"
		logger.Info("tsdb 未启用，回退到 memory 驱动")
	}

	cfg := map[string]any{}
	switch drv {
	case "influxdb":
		cfg = map[string]any{
			"url":    config.GetString(MODULE, "influxdb.url"),
			"token":  config.GetString(MODULE, "influxdb.token"),
			"org":    config.GetString(MODULE, "influxdb.org"),
			"bucket": config.GetString(MODULE, "influxdb.bucket"),
		}
	case "tdengine":
		cfg = map[string]any{
			"host":     config.GetString(MODULE, "tdengine.host"),
			"port":     config.GetInt(MODULE, "tdengine.port"),
			"username": config.GetString(MODULE, "tdengine.username"),
			"password": config.GetString(MODULE, "tdengine.password"),
			"database": config.GetString(MODULE, "tdengine.database"),
		}
	}

	s, err := New(drv, cfg)
	if err != nil {
		return err
	}
	DefaultStore = s
	logger.Info("tsdb 已初始化", "driver", drv)
	return nil
}

func Shutdown() error {
	if DefaultStore != nil {
		err := DefaultStore.Close()
		DefaultStore = nil
		return err
	}
	return nil
}
