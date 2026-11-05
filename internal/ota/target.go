package ota

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"

	"xorm.io/builder"
)

// 未终态的设备升级状态（用于判断设备是否已有进行中的升级任务）
var activeDeviceStatuses = []interface{}{
	types.OTADeviceStatusPending,
	types.OTADeviceStatusDispatched,
	types.OTADeviceStatusDownloading,
	types.OTADeviceStatusVerifying,
	types.OTADeviceStatusInstalling,
	types.OTADeviceStatusRebooting,
}

// ResolveDevices 按目标范围解析设备列表（已排除禁用设备），按创建时间、ID 稳定排序。
func ResolveDevices(scopeType, scopeID string) ([]types.Device, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("请选择升级目标")
	}

	var devices []types.Device
	switch scopeType {
	case types.OTAScopeDevice:
		var dev types.Device
		found, err := db.FindOne("device", builder.Eq{"id": scopeID, "disabled": false}, &dev)
		if err != nil {
			return nil, err
		}
		if !found || dev.ID == "" {
			return nil, fmt.Errorf("设备不存在或已禁用")
		}
		devices = []types.Device{dev}
	case types.OTAScopeModel:
		if err := db.Find("device", builder.Eq{"model_id": scopeID, "disabled": false}, &devices, 0, 0); err != nil {
			return nil, err
		}
	case types.OTAScopeGroup:
		groupID, err := strconv.ParseInt(scopeID, 10, 64)
		if err != nil || groupID <= 0 {
			return nil, fmt.Errorf("无效的分组ID")
		}
		if err := db.Find("device", builder.Eq{"group_id": groupID, "disabled": false}, &devices, 0, 0); err != nil {
			return nil, err
		}
	case types.OTAScopeOrg:
		orgID, err := strconv.ParseInt(scopeID, 10, 64)
		if err != nil || orgID <= 0 {
			return nil, fmt.Errorf("无效的组织ID")
		}
		ids, err := orgSubtreeIDs(orgID)
		if err != nil {
			return nil, err
		}
		deviceIDs, err := orgDeviceIDs(ids)
		if err != nil {
			return nil, err
		}
		if len(deviceIDs) == 0 {
			return nil, nil
		}
		if err := db.Find("device", builder.And(
			builder.In("id", deviceIDs...),
			builder.Eq{"disabled": false},
		), &devices, 0, 0); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("不支持的升级目标类型")
	}

	sort.SliceStable(devices, func(i, j int) bool {
		if devices[i].Created.Equal(devices[j].Created) {
			return devices[i].ID < devices[j].ID
		}
		return devices[i].Created.Before(devices[j].Created)
	})
	return devices, nil
}

// orgSubtreeIDs 返回组织及其全部子组织的 ID
func orgSubtreeIDs(orgID int64) ([]int64, error) {
	var org types.Org
	found, err := db.FindOne("org", builder.Eq{"id": orgID}, &org)
	if err != nil {
		return nil, err
	}
	if !found || org.ID == 0 {
		return nil, fmt.Errorf("组织不存在")
	}

	ids := []int64{org.ID}
	var subs []types.Org
	var cond builder.Cond = builder.Eq{"id": org.ID}
	if org.Path != "" {
		cond = builder.Or(builder.Eq{"id": org.ID}, builder.Like{"path", org.Path + "%"})
	}
	if err := db.Find("org", cond, &subs, 0, 0); err != nil {
		return nil, err
	}
	seen := map[int64]bool{org.ID: true}
	for _, o := range subs {
		if !seen[o.ID] {
			seen[o.ID] = true
			ids = append(ids, o.ID)
		}
	}
	return ids, nil
}

// orgDeviceIDs 查询绑定到指定组织集合的设备 ID（去重）
func orgDeviceIDs(orgIDs []int64) ([]interface{}, error) {
	orgArgs := make([]interface{}, 0, len(orgIDs))
	for _, id := range orgIDs {
		orgArgs = append(orgArgs, id)
	}
	var rels []types.DeviceOrg
	if err := db.Find("device_org", builder.In("org_id", orgArgs...), &rels, 0, 0); err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(rels))
	out := make([]interface{}, 0, len(rels))
	for _, rel := range rels {
		if rel.DeviceID == "" || seen[rel.DeviceID] {
			continue
		}
		seen[rel.DeviceID] = true
		out = append(out, rel.DeviceID)
	}
	return out, nil
}

// FilterCompatible 过滤与固件适用型号不匹配的设备（固件 model_id 为空表示通用）
func FilterCompatible(devices []types.Device, modelID string) []types.Device {
	if modelID == "" {
		return devices
	}
	out := make([]types.Device, 0, len(devices))
	for _, d := range devices {
		if d.ModelID == modelID {
			out = append(out, d)
		}
	}
	return out
}

// ActiveDeviceIDs 返回已有未终态升级任务的设备集合（可排除指定任务）。
func ActiveDeviceIDs(deviceIDs []string, exceptTaskID string) (map[string]bool, error) {
	out := make(map[string]bool, len(deviceIDs))
	if len(deviceIDs) == 0 {
		return out, nil
	}
	args := make([]interface{}, 0, len(deviceIDs))
	for _, id := range deviceIDs {
		if id != "" {
			args = append(args, id)
		}
	}
	if len(args) == 0 {
		return out, nil
	}

	conds := []builder.Cond{
		builder.In("device_id", args...),
		builder.In("status", activeDeviceStatuses...),
	}
	if exceptTaskID != "" {
		conds = append(conds, builder.Neq{"task_id": exceptTaskID})
	}
	var rows []types.OTATaskDevice
	if err := db.Find("ota_task_device", builder.And(conds...), &rows, 0, 0); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.DeviceID] = true
	}
	return out, nil
}

// PrepareStats 目标解析统计（用于预览展示排除原因）
type PrepareStats struct {
	Matched         int  // 型号匹配后的设备数
	Excluded        int  // 已有进行中任务而排除
	WhiteFiltered   int  // 不在白名单而排除
	BlackFiltered   int  // 命中黑名单而排除
	WhitelistActive bool // 固件是否启用白名单
}

// RuleFiltered 因名单规则被排除的设备数
func (s PrepareStats) RuleFiltered() int {
	return s.WhiteFiltered + s.BlackFiltered
}

// PrepareTargets 解析并过滤可升级设备：型号匹配 + 固件黑白名单 + 排除已有进行中任务的设备
func PrepareTargets(fw *types.Firmware, scopeType, scopeID, exceptTaskID string) (devices []types.Device, stats PrepareStats, err error) {
	all, err := ResolveDevices(scopeType, scopeID)
	if err != nil {
		return nil, stats, err
	}
	all = FilterCompatible(all, fw.ModelID)
	stats.Matched = len(all)
	if len(all) == 0 {
		return nil, stats, nil
	}

	// 固件级黑白名单：白名单启用时仅名单内设备通过，黑名单命中始终排除
	rules, err := LoadRuleSet(fw.ID)
	if err != nil {
		return nil, stats, err
	}
	stats.WhitelistActive = rules.WhiteActive
	all, stats.WhiteFiltered, stats.BlackFiltered = rules.Filter(all)
	if len(all) == 0 {
		return nil, stats, nil
	}

	ids := make([]string, 0, len(all))
	for _, d := range all {
		ids = append(ids, d.ID)
	}
	busy, err := ActiveDeviceIDs(ids, exceptTaskID)
	if err != nil {
		return nil, stats, err
	}
	out := make([]types.Device, 0, len(all))
	for _, d := range all {
		if busy[d.ID] {
			stats.Excluded++
			continue
		}
		out = append(out, d)
	}
	return out, stats, nil
}

// terminalDeviceStatus 设备状态是否为终态
func terminalDeviceStatus(status string) bool {
	switch status {
	case types.OTADeviceStatusSuccess,
		types.OTADeviceStatusFailed,
		types.OTADeviceStatusTimeout,
		types.OTADeviceStatusCancelled,
		types.OTADeviceStatusRolledBack:
		return true
	}
	return false
}

// countTaskDevices 统计任务设备的成功数/失败数/终态数
func countTaskDevices(taskID string, batchSeq int) (total, success, failed, terminal int, err error) {
	var rows []types.OTATaskDevice
	var cond builder.Cond = builder.Eq{"task_id": taskID}
	if batchSeq > 0 {
		cond = builder.And(cond, builder.Eq{"batch_seq": batchSeq})
	}
	if err = db.Find("ota_task_device", cond, &rows, 0, 0); err != nil {
		return
	}
	total = len(rows)
	for _, r := range rows {
		if r.Status == types.OTADeviceStatusSuccess {
			success++
			terminal++
			continue
		}
		if r.Status == types.OTADeviceStatusFailed ||
			r.Status == types.OTADeviceStatusTimeout ||
			r.Status == types.OTADeviceStatusRolledBack {
			failed++
			terminal++
			continue
		}
		if r.Status == types.OTADeviceStatusCancelled {
			terminal++
		}
	}
	return
}

// lastActivity 设备明细最近活动时间（用于超时判断）
func lastActivity(row *types.OTATaskDevice) time.Time {
	if !row.Updated.IsZero() {
		return row.Updated
	}
	return row.DispatchedAt
}
