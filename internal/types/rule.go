package types

import (
	"time"
)

// 规则触发类型常量
const (
	RuleTriggerProperty = "property" // 属性上报触发
	RuleTriggerEvent    = "event"    // 设备事件触发
	RuleTriggerStatus   = "status"   // 设备上下线触发
	RuleTriggerSchedule = "schedule" // 定时触发（cron 表达式）
)

// 规则作用范围类型常量
const (
	RuleScopeSingle = "single" // 单个设备
	RuleScopeModel  = "model"  // 某物模型下全部设备
	RuleScopeGroup  = "group"  // 某设备分组下全部设备
)

// 规则条件组合逻辑常量
const (
	RuleLogicAnd = "and" // 条件全部满足
	RuleLogicOr  = "or"  // 任一条件满足
)

// 规则动作类型常量
const (
	RuleActionAlarm           = "alarm"            // 异常报警：写入 device_event
	RuleActionControlProperty = "control_property" // 自动控制：设置设备属性
	RuleActionControlService  = "control_service"  // 自动控制：调用设备服务
	RuleActionNotify          = "notify"           // 通知：写入通知类事件
	RuleActionData            = "data"             // 数据加工：写衍生指标到时序库
)

// 规则告警状态机状态常量
const (
	RuleStateOK       = "ok"       // 正常（条件不满足）
	RuleStatePending  = "pending"  // 去抖中（首次命中，持续观察）
	RuleStateFiring   = "firing"   // 告警中
	RuleStateResolved = "resolved" // 已恢复（瞬时态，落事件后回到 ok）
)

// Rule 自动化规则，描述"在什么范围、由什么触发、满足什么条件、执行什么动作"
type Rule struct {
	ID         string          `json:"id,omitempty" xorm:"id char(32) pk notnull"`              // 规则ID
	Name       string          `json:"name,omitempty" xorm:"name varchar(128)"`                 // 规则名称
	Desc       string          `json:"desc,omitempty" xorm:"desc varchar(512)"`                 // 规则描述
	Disabled   bool            `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Priority   int             `json:"priority,omitempty" xorm:"priority int default 0"`        // 优先级（数值越大越优先）
	Scope      RuleScope       `json:"scope,omitempty" xorm:"scope text json"`                  // 作用范围（JSON）
	Trigger    RuleTrigger     `json:"trigger,omitempty" xorm:"trigger text json"`              // 触发方式（JSON）
	Expr       string          `json:"expr,omitempty" xorm:"expr text"`                         // 表达式条件（expr-lang 语法）
	Conditions []RuleCondition `json:"conditions,omitempty" xorm:"conditions mediumtext json"`  // 结构化条件列表（JSON）
	Logic      string          `json:"logic,omitempty" xorm:"logic varchar(8) default 'and'"`   // 条件组合逻辑（and/or）
	Actions    []RuleAction    `json:"actions,omitempty" xorm:"actions mediumtext json"`        // 命中后执行的动作列表（JSON）
	Debounce   int64           `json:"debounce,omitempty" xorm:"debounce bigint default 0"`     // 去抖时长（秒），0 表示不去抖
	Throttle   int64           `json:"throttle,omitempty" xorm:"throttle bigint default 0"`     // 抑制时长（秒），两次触发的最小间隔
	Created    time.Time       `json:"created,omitempty,zeroempty" xorm:"created datetime"`     // 创建时间
	Updated    time.Time       `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`     // 更新时间
}

// RuleScope 规则作用范围
type RuleScope struct {
	Type     string `json:"type,omitempty"`      // 范围类型（single/model/group）
	DeviceID string `json:"device_id,omitempty"` // 单设备范围时的设备ID
	ModelID  string `json:"model_id,omitempty"`  // 物模型范围时的物模型ID
	GroupID  string `json:"group_id,omitempty"`  // 设备分组范围时的分组ID
}

// RuleTrigger 规则触发方式
type RuleTrigger struct {
	Type      string `json:"type,omitempty"`       // 触发类型（property/event/status/schedule）
	Property  string `json:"property,omitempty"`   // 属性触发时的属性名
	EventName string `json:"event_name,omitempty"` // 事件触发时的事件名
	Status    string `json:"status,omitempty"`     // 状态触发时的状态值（online/offline）
	Cron      string `json:"cron,omitempty"`       // 定时触发时的 cron 表达式
}

// RuleCondition 结构化规则条件
type RuleCondition struct {
	Field string      `json:"field,omitempty"` // 取值字段
	Op    string      `json:"op,omitempty"`    // 比较操作符（如 >、<、==）
	Value interface{} `json:"value,omitempty"` // 比较目标值
	Edge  bool        `json:"edge,omitempty"`  // 是否仅在状态变化（边沿）时触发
}

// RuleAction 规则动作
type RuleAction struct {
	Type   string                 `json:"type,omitempty"`   // 动作类型（alarm/control_property/control_service/notify/data）
	Params map[string]interface{} `json:"params,omitempty"` // 动作参数
}

// RuleLog 规则执行日志
// ID 使用 int64 自增；RuleID→Rule(string), DeviceID→Device(string)
type RuleLog struct {
	ID       int64                  `json:"id,omitempty" xorm:"id bigint pk autoincr"`             // 日志ID
	RuleID   string                 `json:"rule_id,omitempty" xorm:"rule_id char(32) index"`       // 规则ID → Rule(string)
	RuleName string                 `json:"rule_name,omitempty" xorm:"rule_name varchar(128)"`     // 规则名称快照
	DeviceID string                 `json:"device_id,omitempty" xorm:"device_id char(32) index"`   // 设备ID → Device(string)
	Trigger  string                 `json:"trigger,omitempty" xorm:"trigger varchar(32)"`          // 本次触发类型
	Values   map[string]interface{} `json:"values,omitempty" xorm:"values text json"`              // 触发时的变量值（JSON）
	Matched  bool                   `json:"matched,omitempty" xorm:"matched tinyint(1) default 0"` // 条件是否命中
	State    string                 `json:"state,omitempty" xorm:"state varchar(32)"`              // 告警状态机状态（ok/pending/firing/resolved）
	Results  []RuleActionResult     `json:"results,omitempty" xorm:"results text json"`            // 各动作执行结果（JSON）
	Created  time.Time              `json:"created,omitempty,zeroempty" xorm:"created datetime"`   // 记录时间
}

// RuleActionResult 单个动作的执行结果
type RuleActionResult struct {
	Type    string `json:"type,omitempty"`    // 动作类型
	Success bool   `json:"success,omitempty"` // 是否执行成功
	Message string `json:"message,omitempty"` // 结果消息或失败原因
}
