package tsdb

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper 模块名
const MODULE = "tsdb"

func init() {
	config.SetDefault(MODULE, "enable", false)
	config.SetDefault(MODULE, "driver", "memory")

	// driver=influxdb 时读
	config.SetDefault(MODULE, "influxdb.url", "")
	config.SetDefault(MODULE, "influxdb.token", "")
	config.SetDefault(MODULE, "influxdb.org", "")
	config.SetDefault(MODULE, "influxdb.bucket", "")

	// driver=tdengine 时读
	config.SetDefault(MODULE, "tdengine.host", "")
	config.SetDefault(MODULE, "tdengine.port", 6041)
	config.SetDefault(MODULE, "tdengine.username", "root")
	config.SetDefault(MODULE, "tdengine.password", "taosdata")
	config.SetDefault(MODULE, "tdengine.database", "dmp")
}

func Enabled() bool { return config.GetBool(MODULE, "enable") }
func Driver() string {
	d := config.GetString(MODULE, "driver")
	if d == "" {
		return "memory"
	}
	return d
}
