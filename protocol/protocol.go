package protocol

import (
	"github.com/busy-cloud/boat/smart"
)

type Base struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version,omitempty"`
	Author      string `json:"author,omitempty"`
	Copyright   string `json:"copyright,omitempty"`
}

type Protocol struct {
	Base

	Station []*smart.Field `json:"station,omitempty"` //从站信息
	Options []*smart.Field `json:"options,omitempty"` //协议参数
	Model   []*smart.Field `json:"model,omitempty"`   //模型配置文件
}
