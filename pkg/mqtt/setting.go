package mqtt

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/smart"
)

func init() {
	config.Register(MODULE, &config.Form{
		Title:  "MQTT连接配置",
		Module: MODULE,
		Fields: []smart.Field{
			{Key: "url", Label: "地址", Type: "text", Required: true, Default: ""},
			{Key: "clientid", Label: "客户端ID", Type: "text"},
			{Key: "username", Label: "用户名", Type: "text"},
			{Key: "password", Label: "密码", Type: "text"},
		},
	})
}
