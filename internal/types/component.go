package types

import (
	"time"
)

// ComponentVariable 组件变量定义
type ComponentVariable struct {
	Key         string      `json:"key,omitempty"`         // 变量键名
	Name        string      `json:"name,omitempty"`        // 变量显示名
	Type        string      `json:"type,omitempty"`        // 变量数据类型
	Default     interface{} `json:"default,omitempty"`     // 变量默认值
	Description string      `json:"description,omitempty"` // 变量说明
}

// ComponentEvent 组件事件定义
type ComponentEvent struct {
	Key         string `json:"key,omitempty"`         // 事件键名
	Name        string `json:"name,omitempty"`        // 事件显示名
	Description string `json:"description,omitempty"` // 事件说明
}

// ComponentScript 组件脚本定义
type ComponentScript struct {
	Name string `json:"name,omitempty"` // 脚本名称
	Code string `json:"code,omitempty"` // 脚本代码
}

// Component 组态组件定义
type Component struct {
	ID          string              `json:"id,omitempty" xorm:"id char(32) pk"`                      // 组件ID
	Name        string              `json:"name,omitempty" xorm:"name varchar(128)"`                 // 组件名称
	Description string              `json:"description,omitempty" xorm:"description varchar(512)"`   // 组件说明
	Icon        string              `json:"icon,omitempty" xorm:"icon varchar(512)"`                 // 组件图标
	Category    string              `json:"category,omitempty" xorm:"category varchar(64)"`          // 所属分类（basic / control）
	Elements    []interface{}       `json:"elements" xorm:"elements mediumtext json"`                // 图元数据（JSON）
	Variables   []ComponentVariable `json:"variables,omitempty" xorm:"variables text json"`          // 变量定义列表（JSON）
	Events      []ComponentEvent    `json:"events,omitempty" xorm:"events text json"`                // 事件定义列表（JSON）
	Scripts     []ComponentScript   `json:"scripts,omitempty" xorm:"scripts mediumtext json"`        // 脚本定义列表（JSON）
	Disabled    bool                `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Created     time.Time           `json:"created,omitempty" xorm:"created datetime"`               // 创建时间
	Updated     time.Time           `json:"updated,omitempty" xorm:"updated datetime"`               // 更新时间
}

// 组件分组常量
const (
	ComponentCategoryBasic   = "basic"   // 基础图元分类
	ComponentCategoryControl = "control" // 页面控件分类
)

// ComponentCategoryNames 组件分类到中文名称的映射
var ComponentCategoryNames = map[string]string{
	ComponentCategoryBasic:   "基础图元",
	ComponentCategoryControl: "页面控件",
}
