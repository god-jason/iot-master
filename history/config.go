package history

import (
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/smart"
)

const MODULE = "influxdb"

func init() {
	config.SetDefault(MODULE, "enable", false)
	config.SetDefault(MODULE, "url", "http://127.0.0.1:8086")
	config.SetDefault(MODULE, "org", "")
	config.SetDefault(MODULE, "bucket", "")
	config.SetDefault(MODULE, "token", "")

	config.Register(MODULE, &config.Form{
		Title:  "Influxdb数据库配置",
		Module: MODULE,
		Fields: []smart.Field{
			{Key: "enable", Label: "启用", Type: "switch"},
			{Key: "url", Label: "服务器地址", Type: "text", Default: "http://127.0.0.1:8086"},
			{Key: "org", Label: "Org", Type: "text"},
			{Key: "bucket", Label: "Bucket", Type: "text"},
			{Key: "token", Label: "Token", Type: "text"},
		},
	})
}
