package oem

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/smart"
)

func init() {
	config.Register(MODULE, &config.Form{
		Title:  "OEM配置",
		Module: MODULE,
		Fields: []smart.Field{
			{Key: "name", Label: "名称", Type: "text"},
			{Key: "logo", Label: "图标", Type: "text"},
		},
	})
}
