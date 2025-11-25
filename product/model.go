package product

import (
	"time"

	"github.com/busy-cloud/boat/smart"
)

// Point 属性
type Point struct {
	Name      string `json:"name,omitempty"`  //变量名称
	Label     string `json:"label,omitempty"` //显示名称
	Unit      string `json:"unit,omitempty"`  //单位
	Type      string `json:"type,omitempty"`  //bool string number array object
	Mode      string `json:"mode,omitempty"`  //模式，rw r w
	Precision uint8  `json:"precision,omitempty"`
	Default   any    `json:"default,omitempty"` //默认值
	//Writable  bool   `json:"writable,omitempty"` //是否可写
	//History   bool   `json:"history,omitempty"`  //是否保存历史
}

type Property struct {
	Name   string           `json:"name,omitempty"`
	Points []map[string]any `json:"points,omitempty"` //要支持扩展字段，所以用map数组
}

type Parameter struct {
	Key   string `json:"key,omitempty"`
	Label string `json:"label,omitempty"`
	Type  string `json:"type,omitempty"`
}

type Event struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

type Action struct {
	Name        string      `json:"name,omitempty"`
	Label       string      `json:"label,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
	Returns     []Parameter `json:"returns,omitempty"`
}

type Operator struct {
	Name   string `json:"name,omitempty"`
	Label  string `json:"label,omitempty"`
	Type   string `json:"type,omitempty"`   //类型 button switch/toggle slider
	Bind   string `json:"bind,omitempty"`   //绑定值(状态)
	Action string `json:"action,omitempty"` //操作
}

type Setting struct {
	Name   string        `json:"name,omitempty"`
	Label  string        `json:"label,omitempty"`
	Fields []smart.Field `json:"fields,omitempty"`
}
type ProductModel struct {
	Id         string       `json:"id,omitempty" xorm:"pk"`
	Properties []*Property  `json:"properties,omitempty" xorm:"json"` //属性表，分组的形式
	Events     []*Event     `json:"events,omitempty" xorm:"json"`     //事件表
	Actions    []*Action    `json:"actions,omitempty" xorm:"json"`    //响应表
	Validators []*Validator `json:"validators,omitempty" xorm:"json"` //检查
	Operators  []*Operator  `json:"operators,omitempty" xorm:"json"`  //按钮操作
	Settings   []*Setting   `json:"settings,omitempty" xorm:"json"`   //参数配置
	Updated    time.Time    `json:"updated,omitempty" xorm:"updated"`
	Created    time.Time    `json:"created,omitempty" xorm:"created"`
}
