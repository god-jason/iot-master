package product

import (
	"time"
)

// Property 属性
type Property struct {
	Name      string `json:"name,omitempty"`  //变量名称
	Label     string `json:"label,omitempty"` //显示名称
	Unit      string `json:"unit,omitempty"`  //单位
	Type      string `json:"type,omitempty"`  //bool string number array object
	Precision uint8  `json:"precision,omitempty"`
	Default   any    `json:"default,omitempty"`  //默认值
	Writable  bool   `json:"writable,omitempty"` //是否可写
	History   bool   `json:"history,omitempty"`  //是否保存历史
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

type Model struct {
	Id         string       `json:"id,omitempty" xorm:"pk"`
	Properties []*Property  `json:"properties,omitempty" xorm:"json"`
	Events     []*Event     `json:"events,omitempty" xorm:"json"`
	Actions    []*Action    `json:"actions,omitempty" xorm:"json"`
	Validators []*Validator `json:"validators,omitempty" xorm:"json"`
	Updated    time.Time    `json:"updated,omitempty" xorm:"updated"`
	Created    time.Time    `json:"created,omitempty" xorm:"created"`
}
