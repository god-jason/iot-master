package product

import (
	"github.com/busy-cloud/boat/db"
	"time"
)

func init() {
	db.Register(&Product{}, &ProductConfig{}, &Model{})
}

type Product struct {
	Id          string    `json:"id,omitempty" xorm:"pk"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type,omitempty"` //类型
	Version     string    `json:"version,omitempty"`
	Protocol    string    `json:"protocol,omitempty"`
	Disabled    bool      `json:"disabled,omitempty"` //禁用
	Created     time.Time `json:"created,omitempty" xorm:"created"`
}

type ProductConfig struct {
	Id      string         `json:"id" xorm:"pk"`
	Name    string         `json:"name" xorm:"pk"` //双主键
	Content map[string]any `json:"content,omitempty" xorm:"json text"`
	Updated time.Time      `json:"updated,omitempty" xorm:"updated"`
	Created time.Time      `json:"created,omitempty" xorm:"created"`
}
