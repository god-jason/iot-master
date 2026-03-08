package web

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/smart"
)

func init() {
	config.Register(MODULE, &config.Form{
		Title:  "Web配置",
		Module: MODULE,
		Fields: []smart.Field{
			{
				Key: "mode", Label: "HTTPS", Type: "select",
				Options: []smart.SelectOption{
					{Label: "HTTP", Value: "HTTP"},
					{Label: "TLS", Value: "TLS"},
					{Label: "LetsEncrypt", Value: "LetsEncrypt"},
				},
			},
			{Key: "port", Label: "端口", Type: "number", Required: true, Default: 8080, Min: 1, Max: 65535},
			{Key: "debug", Label: "调试模式", Type: "switch"},
			{Key: "cors", Label: "跨域请求", Type: "switch"},
			{Key: "gzip", Label: "压缩模式", Type: "switch"},
			{Key: "tls_cert", Label: "证书cert", Type: "file"},
			{Key: "tls_key", Label: "证书key", Type: "file"},
			{Key: "email", Label: "E-Mail", Type: "text"},
			{Key: "hosts", Label: "域名", Type: "tags", Default: []string{}},
		},
	})
}
