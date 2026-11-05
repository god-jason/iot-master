package types

import (
	"time"
)

// Task 定时任务：按 5 段 cron 表达式定时触发一个场景
// ID 使用 int64 自增；SceneID→Scene(int64)
type Task struct {
	ID       int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`               // 任务ID
	Name     string    `json:"name,omitempty" xorm:"name varchar(128)"`                 // 任务名称
	Desc     string    `json:"desc,omitempty" xorm:"desc varchar(512)"`                 // 任务描述
	Mode     string    `json:"mode,omitempty" xorm:"mode varchar(16)"`                  // 运行模式（online/offline）
	Disabled bool      `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Cron     string    `json:"cron,omitempty" xorm:"cron varchar(64)"`                  // 5 段 cron 表达式
	SceneID  int64     `json:"scene_id,omitempty" xorm:"scene_id bigint index"`         // 触发的场景ID → Scene(int64)
	LastRun  time.Time `json:"last_run,omitempty,zeroempty" xorm:"last_run datetime"`   // 上次执行时间
	Created  time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`     // 创建时间
	Updated  time.Time `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`     // 更新时间
}
