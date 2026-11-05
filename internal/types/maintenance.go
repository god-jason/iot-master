// Package types 维保管理相关数据模型。
//
// 两张表：
//   - maintenance_plan  维保计划（按型号/分组/组织/设备范围周期性生成工单）
//   - maintenance_order 维保工单（计划生成或手工创建，记录执行与完成信息）
package types

import "time"

// 维保目标范围类型
const (
	MaintenanceScopeModel  = "model"  // 按物模型
	MaintenanceScopeGroup  = "group"  // 按设备分组
	MaintenanceScopeOrg    = "org"    // 按组织（含子组织）
	MaintenanceScopeDevice = "device" // 指定设备
)

// 维保周期类型
const (
	MaintenanceCycleDay   = "day"   // 按天
	MaintenanceCycleMonth = "month" // 按月
	MaintenanceCycleYear  = "year"  // 按年
)

// 维保工单状态
const (
	MaintenanceOrderStatusPending    = "pending"    // 待处理
	MaintenanceOrderStatusProcessing = "processing" // 处理中
	MaintenanceOrderStatusReviewing  = "reviewing"  // 待复核
	MaintenanceOrderStatusCompleted  = "completed"  // 已完成
	MaintenanceOrderStatusCancelled  = "cancelled"  // 已取消
)

// 维保工单类型
const (
	MaintenanceOrderTypePlan    = "plan"    // 计划保养
	MaintenanceOrderTypeFault   = "fault"   // 故障维修
	MaintenanceOrderTypePatrol  = "patrol"  // 巡检
	MaintenanceOrderTypeInstall = "install" // 安装调试
)

// 维保工单结果
const (
	MaintenanceResultNormal     = "normal"     // 正常
	MaintenanceResultFixed      = "fixed"      // 已修复
	MaintenanceResultReplaced   = "replaced"   // 已更换
	MaintenanceResultUnresolved = "unresolved" // 未解决
)

// MaintenancePlan 维保计划实体
type MaintenancePlan struct {
	ID            int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`                         // 计划ID
	Name          string    `json:"name,omitempty" xorm:"name varchar(128)"`                           // 计划名称
	ScopeType     string    `json:"scope_type,omitempty" xorm:"scope_type varchar(16)"`                // 目标范围类型：model / group / org / device
	ScopeID       string    `json:"scope_id,omitempty" xorm:"scope_id varchar(64)"`                    // 目标范围ID
	CycleType     string    `json:"cycle_type,omitempty" xorm:"cycle_type varchar(16)"`                // 周期类型：day / month / year
	CycleValue    int       `json:"cycle_value,omitempty" xorm:"cycle_value int default 1"`            // 周期值（每 N 天/月/年）
	RemindDays    int       `json:"remind_days,omitempty" xorm:"remind_days int default 0"`            // 提前提醒天数（到期前 N 天生成工单）
	Items         string    `json:"items,omitempty" xorm:"items text"`                                 // 检查项模板 JSON
	OwnerID       string    `json:"owner_id,omitempty" xorm:"owner_id char(32) index"`                 // 默认负责人用户ID
	NextDue       time.Time `json:"next_due,omitempty,zeroempty" xorm:"next_due datetime index"`       // 下次到期时间
	LastGenerated time.Time `json:"last_generated,omitempty,zeroempty" xorm:"last_generated datetime"` // 最近生成时间
	Disabled      bool      `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"`           // 是否停用
	CreatedBy     string    `json:"created_by,omitempty" xorm:"created_by char(32)"`                   // 创建人
	Created       time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`               // 创建时间
	Updated       time.Time `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`               // 更新时间
}

// TableName 维保计划表名
func (MaintenancePlan) TableName() string { return "maintenance_plan" }

// MaintenanceOrder 维保工单实体
type MaintenanceOrder struct {
	ID          int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`                   // 工单ID
	OrderNo     string    `json:"order_no,omitempty" xorm:"order_no varchar(32) unique"`       // 工单号（唯一）
	PlanID      int64     `json:"plan_id,omitempty" xorm:"plan_id bigint index"`               // 关联维保计划ID，0 表示手工工单
	DeviceID    string    `json:"device_id,omitempty" xorm:"device_id char(32) index"`         // 关联设备ID
	Type        string    `json:"type,omitempty" xorm:"type varchar(16)"`                      // 工单类型：plan / fault / patrol / install
	Title       string    `json:"title,omitempty" xorm:"title varchar(128)"`                   // 工单标题
	Priority    int       `json:"priority" xorm:"priority int default 2"`                      // 优先级：1 高 / 2 中 / 3 低
	Status      string    `json:"status,omitempty" xorm:"status varchar(16) index"`            // 工单状态
	AssigneeID  string    `json:"assignee_id,omitempty" xorm:"assignee_id char(32) index"`     // 负责人用户ID
	DueAt       time.Time `json:"due_at,omitempty,zeroempty" xorm:"due_at datetime index"`     // 应完成时间
	StartedAt   time.Time `json:"started_at,omitempty,zeroempty" xorm:"started_at datetime"`   // 开始处理时间
	FinishedAt  time.Time `json:"finished_at,omitempty,zeroempty" xorm:"finished_at datetime"` // 完成时间
	Result      string    `json:"result,omitempty" xorm:"result varchar(16)"`                  // 处理结果：normal / fixed / replaced / unresolved
	Items       string    `json:"items,omitempty" xorm:"items text"`                           // 检查项 JSON（执行时填充结果）
	Attachments string    `json:"attachments,omitempty" xorm:"attachments text"`               // 附件 JSON
	Parts       string    `json:"parts,omitempty" xorm:"parts text"`                           // 更换备件 JSON
	Cost        float64   `json:"cost" xorm:"cost decimal(10,2) default 0"`                    // 费用
	Remark      string    `json:"remark,omitempty" xorm:"remark varchar(512)"`                 // 备注
	CreatedBy   string    `json:"created_by,omitempty" xorm:"created_by char(32)"`             // 创建人
	Created     time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`         // 创建时间
	Updated     time.Time `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`         // 更新时间
}

// TableName 维保工单表名
func (MaintenanceOrder) TableName() string { return "maintenance_order" }
