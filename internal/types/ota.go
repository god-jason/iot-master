// Package types OTA 固件升级相关数据模型。
//
// 四张表：
//   - firmware        固件本体（文件、SHA256、适用型号、发布状态）
//   - ota_task        升级任务（目标范围、灰度策略、定时、汇总统计）
//   - ota_task_batch  灰度批次（百分比、等待时间、批次状态）
//   - ota_task_device 任务设备明细（下发状态、进度、重试次数）
package types

import "time"

// 固件状态
const (
	FirmwareStatusDraft    = 0 // 草稿（仅上传，不可下发）
	FirmwareStatusReleased = 1 // 已发布，可用于升级任务
	FirmwareStatusDisabled = 2 // 已停用
)

// Firmware 固件实体
type Firmware struct {
	ID         string    `json:"id,omitempty" xorm:"id char(32) pk notnull"`            // 固件ID
	ModelID    string    `json:"model_id,omitempty" xorm:"model_id char(32) index"`     // 适用物模型ID，空表示通用固件
	Name       string    `json:"name,omitempty" xorm:"name varchar(128)"`               // 固件名称
	Version    string    `json:"version,omitempty" xorm:"fw_version varchar(64) index"` // 版本号（列名避开 xorm 的 version 乐观锁语义）
	FileKey    string    `json:"file_key,omitempty" xorm:"file_key varchar(255)"`       // 磁盘文件名（存储目录内）
	FileURL    string    `json:"file_url,omitempty" xorm:"file_url varchar(512)"`       // 下载地址（相对路径 /api/ota/firmware/{id}/file）
	FileSize   int64     `json:"file_size,omitempty" xorm:"file_size bigint"`           // 文件大小（字节）
	SHA256     string    `json:"sha256,omitempty" xorm:"sha256 varchar(64)"`            // 文件 SHA256
	Signature  string    `json:"signature,omitempty" xorm:"signature varchar(512)"`     // 固件签名（可选）
	MinVersion string    `json:"min_version,omitempty" xorm:"min_version varchar(64)"`  // 最低可升级版本
	Channel    string    `json:"channel,omitempty" xorm:"channel varchar(16)"`          // 通道：stable / beta
	Notes      string    `json:"notes,omitempty" xorm:"notes text"`                     // 备注
	Status     int       `json:"status,omitempty" xorm:"status int default 0"`          // 状态：0草稿 1已发布 2已停用
	CreatedBy  string    `json:"created_by,omitempty" xorm:"created_by char(32) index"` // 创建人
	Created    time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`   // 创建时间
	Updated    time.Time `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`   // 更新时间
}

// TableName 固件表名
// 表名使用 ota_firmware，避免与历史遗留的 firmware 表冲突
func (Firmware) TableName() string { return "ota_firmware" }

// 固件名单类型
const (
	FirmwareListWhite = "white" // 白名单：存在白名单规则时，仅名单内设备允许升级
	FirmwareListBlack = "black" // 黑名单：名单内设备禁止升级（优先级高于白名单）
)

// 名单目标类型
const (
	RuleTargetDevice = "device" // 具体设备
	RuleTargetGroup  = "group"  // 设备分组（该分组下的设备）
	RuleTargetOrg    = "org"    // 组织（含全部子组织绑定的设备）
)

// FirmwareRule 固件级黑白名单规则
// 规则按固件生效，在预览/创建升级任务时解析为设备集合：
//   - 白名单非空时，只有白名单命中的设备可升级；
//   - 黑名单命中的设备始终排除，两者冲突时以黑名单为准。
type FirmwareRule struct {
	ID         int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`               // 规则ID
	FirmwareID string    `json:"firmware_id,omitempty" xorm:"firmware_id char(32) index"` // 所属固件
	ListType   string    `json:"list_type,omitempty" xorm:"list_type varchar(8) index"`   // 名单类型：white / black
	TargetType string    `json:"target_type,omitempty" xorm:"target_type varchar(16)"`    // 目标类型：device / group / org
	TargetID   string    `json:"target_id,omitempty" xorm:"target_id varchar(64)"`        // 目标ID
	CreatedBy  string    `json:"created_by,omitempty" xorm:"created_by char(32)"`         // 添加人
	Created    time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`     // 添加时间
}

// TableName 固件名单规则表名
func (FirmwareRule) TableName() string { return "ota_firmware_rule" }

// 升级任务状态
const (
	OTATaskStatusScheduled = "scheduled" // 待执行（已创建，未到执行时间）
	OTATaskStatusRunning   = "running"   // 执行中
	OTATaskStatusPaused    = "paused"    // 已暂停
	OTATaskStatusCompleted = "completed" // 全部成功
	OTATaskStatusPartial   = "partial"   // 部分成功
	OTATaskStatusFailed    = "failed"    // 全部失败
	OTATaskStatusCancelled = "cancelled" // 已取消
)

// 升级目标范围
const (
	OTAScopeModel  = "model"  // 按物模型
	OTAScopeGroup  = "group"  // 按设备分组
	OTAScopeOrg    = "org"    // 按组织（含子组织）
	OTAScopeDevice = "device" // 指定设备
)

// OTATask 升级任务
type OTATask struct {
	ID           string    `json:"id,omitempty" xorm:"id char(32) pk notnull"`                        // 任务ID
	Name         string    `json:"name,omitempty" xorm:"name varchar(128)"`                           // 任务名称
	FirmwareID   string    `json:"firmware_id,omitempty" xorm:"firmware_id char(32) index"`           // 固件ID
	ScopeType    string    `json:"scope_type,omitempty" xorm:"scope_type varchar(16)"`                // 目标范围类型
	ScopeID      string    `json:"scope_id,omitempty" xorm:"scope_id varchar(64)"`                    // 目标范围ID
	Strategy     string    `json:"strategy,omitempty" xorm:"strategy text"`                           // 灰度策略 JSON
	PauseReason  string    `json:"pause_reason,omitempty" xorm:"pause_reason varchar(32)"`            // 暂停原因：manual / auto
	ScheduledAt  time.Time `json:"scheduled_at,omitempty,zeroempty" xorm:"scheduled_at datetime"`     // 计划执行时间，零值表示立即
	Status       string    `json:"status,omitempty" xorm:"status varchar(16) index"`                  // 任务状态
	Total        int       `json:"total,omitempty" xorm:"total int default 0"`                        // 总设备数
	Success      int       `json:"success,omitempty" xorm:"success int default 0"`                    // 成功数
	Failed       int       `json:"failed,omitempty" xorm:"failed int default 0"`                      // 失败数
	AutoRollback bool      `json:"auto_rollback,omitempty" xorm:"auto_rollback tinyint(1) default 0"` // 失败自动回滚
	CreatedBy    string    `json:"created_by,omitempty" xorm:"created_by char(32) index"`             // 创建人
	StartedAt    time.Time `json:"started_at,omitempty,zeroempty" xorm:"started_at datetime"`         // 开始时间
	FinishedAt   time.Time `json:"finished_at,omitempty,zeroempty" xorm:"finished_at datetime"`       // 结束时间
	Created      time.Time `json:"created,omitempty,zeroempty" xorm:"created datetime"`               // 创建时间
}

// TableName 升级任务表名
func (OTATask) TableName() string { return "ota_task" }

// 批次状态
const (
	OTABatchStatusPending   = "pending"   // 等待上一批完成
	OTABatchStatusRunning   = "running"   // 下发中
	OTABatchStatusCompleted = "completed" // 本批全部终态
)

// OTATaskBatch 灰度批次
type OTATaskBatch struct {
	ID          int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`                   // 批次ID
	TaskID      string    `json:"task_id,omitempty" xorm:"task_id char(32) index"`             // 所属任务
	Seq         int       `json:"seq,omitempty" xorm:"seq int"`                                // 批次序号（从1开始）
	Percent     int       `json:"percent,omitempty" xorm:"percent int"`                        // 目标百分比（0表示重试批次）
	WaitMinutes int       `json:"wait_minutes,omitempty" xorm:"wait_minutes int"`              // 上一批完成后等待分钟数
	Status      string    `json:"status,omitempty" xorm:"status varchar(16)"`                  // 批次状态
	StartedAt   time.Time `json:"started_at,omitempty,zeroempty" xorm:"started_at datetime"`   // 开始时间
	FinishedAt  time.Time `json:"finished_at,omitempty,zeroempty" xorm:"finished_at datetime"` // 完成时间
}

// TableName 灰度批次表名
func (OTATaskBatch) TableName() string { return "ota_task_batch" }

// 单设备升级状态
const (
	OTADeviceStatusPending     = "pending"     // 等待下发（离线或等待批次）
	OTADeviceStatusDispatched  = "dispatched"  // 已下发并收到设备 ACK
	OTADeviceStatusDownloading = "downloading" // 下载固件中
	OTADeviceStatusVerifying   = "verifying"   // 校验固件中
	OTADeviceStatusInstalling  = "installing"  // 写入固件中
	OTADeviceStatusRebooting   = "rebooting"   // 重启中
	OTADeviceStatusSuccess     = "success"     // 升级成功
	OTADeviceStatusFailed      = "failed"      // 升级失败
	OTADeviceStatusTimeout     = "timeout"     // 超时
	OTADeviceStatusCancelled   = "cancelled"   // 已取消
	OTADeviceStatusRolledBack  = "rolled_back" // 已回滚
)

// OTATaskDevice 任务设备明细
type OTATaskDevice struct {
	ID           int64     `json:"id,omitempty" xorm:"id bigint pk autoincr"`                       // 明细ID
	TaskID       string    `json:"task_id,omitempty" xorm:"task_id char(32) index"`                 // 所属任务
	BatchSeq     int       `json:"batch_seq,omitempty" xorm:"batch_seq int"`                        // 所属批次序号
	DeviceID     string    `json:"device_id,omitempty" xorm:"device_id char(32) index"`             // 设备ID
	FromVersion  string    `json:"from_version,omitempty" xorm:"from_version varchar(64)"`          // 升级前版本
	ToVersion    string    `json:"to_version,omitempty" xorm:"to_version varchar(64)"`              // 目标版本
	Status       string    `json:"status,omitempty" xorm:"status varchar(16) index"`                // 升级状态
	Progress     int       `json:"progress,omitempty" xorm:"progress int default 0"`                // 进度 0-100
	RetryCount   int       `json:"retry_count,omitempty" xorm:"retry_count int default 0"`          // 已重试次数
	Error        string    `json:"error,omitempty" xorm:"error varchar(512)"`                       // 失败原因
	DispatchedAt time.Time `json:"dispatched_at,omitempty,zeroempty" xorm:"dispatched_at datetime"` // 下发时间
	FinishedAt   time.Time `json:"finished_at,omitempty,zeroempty" xorm:"finished_at datetime"`     // 终态时间
	Updated      time.Time `json:"updated,omitempty,zeroempty" xorm:"updated datetime"`             // 最近上报时间
}

// TableName 任务设备明细表名
func (OTATaskDevice) TableName() string { return "ota_task_device" }
