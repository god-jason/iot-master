package web

import (
	"github.com/god-jason/iot-master/setting"
	"github.com/god-jason/iot-master/types"
)

func init() {
	setting.Register(MODULE, &setting.Module{
		Name:   "Web",
		Module: MODULE,
		Title:  "Web配置",
		Form: []types.SmartField{
			{Key: "port", Label: "端口", Type: "number", Required: true, Default: 8080, Min: 1, Max: 65535},
			{Key: "debug", Label: "调试模式", Type: "switch"},
			{Key: "cors", Label: "跨域请求", Type: "switch"},
			{Key: "gzip", Label: "压缩模式", Type: "switch"},
			{
				Key: "https", Label: "HTTPS", Type: "select",
				Options: []types.SmartSelectOption{
					{Label: "禁用", Value: ""},
					{Label: "TLS", Value: "TLS"},
					{Label: "LetsEncrypt", Value: "LetsEncrypt"},
				},
			},
			{Key: "cert", Label: "证书cert", Type: "file"},
			{Key: "Key", Label: "证书key", Type: "file"},
			{Key: "email", Label: "E-Mail", Type: "text"},
			{Key: "hosts", Label: "域名", Type: "tags", Default: []string{}},
			{Key: "url", Label: "地址", Type: "text", Required: true, Default: ""},
			{Key: "client_id", Label: "客户端ID", Type: "text"},
			{Key: "username", Label: "用户名", Type: "text"},
			{Key: "password", Label: "密码", Type: "text"},
		},
	})
}
