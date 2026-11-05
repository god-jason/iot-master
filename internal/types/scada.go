package types

import (
	"time"
)

// Scada 组态/大屏画面定义
// type 为 scada 表示组态编辑器画面，为 screen 表示大屏编辑器画面
// 画面内容以 JSON 形式存储在 canvas 与 elements 字段中，由前端画布引擎渲染
type Scada struct {
	ID          int64                  `json:"id,omitempty" xorm:"id bigint pk autoincr"`               // 画面ID
	Name        string                 `json:"name,omitempty" xorm:"name varchar(128)"`                 // 画面名称
	Description string                 `json:"description,omitempty" xorm:"description varchar(512)"`   // 画面描述
	Icon        string                 `json:"icon,omitempty" xorm:"icon varchar(512)"`                 // 画面图标
	Type        string                 `json:"type,omitempty" xorm:"type varchar(16)"`                  // 画面类型（scada 组态 / screen 大屏）
	Canvas      map[string]interface{} `json:"canvas,omitempty" xorm:"canvas mediumtext json"`          // 画布配置（JSON：尺寸、背景等）
	Elements    []interface{}          `json:"elements" xorm:"elements mediumtext json"`                // 画布元素列表（JSON）
	Disabled    bool                   `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Created     time.Time              `json:"created,omitempty" xorm:"created datetime"`               // 创建时间
	Updated     time.Time              `json:"updated,omitempty" xorm:"updated datetime"`               // 更新时间
}
