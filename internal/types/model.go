package types

import (
	"github.com/god-jason/iot-master/pkg/thing"
	"time"
)

// Model 物模型（DB 主表）
// 子结构 Property/Event/Service/... 定义在 pkg/thing 包，这里 JSON 列直接引用
type Model struct {
	ID     string `json:"id,omitempty" xorm:"id char(32) pk notnull"`  // 物模型ID
	Secret string `json:"secret,omitempty" xorm:"secret varchar(255)"` // 物模型密钥（设备接入鉴权用）
	Name   string `json:"name,omitempty" xorm:"name varchar(128)"`     // 物模型名称
	Icon   string `json:"icon,omitempty" xorm:"icon varchar(512)"`     // 物模型图标
	Type   string `json:"type,omitempty" xorm:"type varchar(32)"`      // 物模型类型

	thing.Model `xorm:"extends"` //继承物模型定义

	Created time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"` // 创建时间
}
