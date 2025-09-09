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

	DeviceExtendColumns []*smart.Column `json:"device_extend_columns"` //设备表扩展字段，比如从站号
	DeviceExtendFields  []*smart.Field  `json:"device_extend_fields"`  //设备编辑扩展字段，比如从站号

	PointExtendFields []*smart.Field `json:"point_extend_fields"` //属性扩展字段 物模型

	OptionFields []*smart.Field `json:"option_fields"` //参数字段（配置在通道之上）
	//Model []*smart.Field `json:"model,omitempty"` //模型配置文件
}
