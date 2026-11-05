package ota

import (
	"fmt"
	"strconv"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"

	"xorm.io/builder"
)

// RuleSet 固件黑白名单解析后的设备集合
// WhiteActive 表示该固件存在白名单规则（即使解析结果为空，也视为启用白名单）
type RuleSet struct {
	White       map[string]bool
	Black       map[string]bool
	WhiteActive bool
}

// LoadRuleSet 读取并解析固件的黑白名单规则为设备 ID 集合
// 目标（设备/分组/组织）已被删除时跳过该条规则，避免历史规则阻塞升级
func LoadRuleSet(firmwareID string) (*RuleSet, error) {
	rs := &RuleSet{White: map[string]bool{}, Black: map[string]bool{}}
	if firmwareID == "" {
		return rs, nil
	}

	var rules []types.FirmwareRule
	if err := db.Find("ota_firmware_rule", builder.Eq{"firmware_id": firmwareID}, &rules, 0, 0); err != nil {
		return nil, err
	}

	for _, r := range rules {
		devices, err := RuleTargetDeviceIDs(r.TargetType, r.TargetID)
		if err != nil {
			continue
		}
		switch r.ListType {
		case types.FirmwareListWhite:
			rs.WhiteActive = true
			for _, id := range devices {
				rs.White[id] = true
			}
		case types.FirmwareListBlack:
			for _, id := range devices {
				rs.Black[id] = true
			}
		}
	}
	return rs, nil
}

// Allowed 判断设备是否允许使用该固件升级（黑名单优先）
func (rs *RuleSet) Allowed(deviceID string) bool {
	if rs == nil {
		return true
	}
	if rs.Black[deviceID] {
		return false
	}
	if rs.WhiteActive {
		return rs.White[deviceID]
	}
	return true
}

// Filter 按名单过滤设备，返回保留设备与白名单/黑名单排除数
func (rs *RuleSet) Filter(devices []types.Device) (kept []types.Device, whiteSkipped, blackSkipped int) {
	kept = make([]types.Device, 0, len(devices))
	for _, d := range devices {
		if rs != nil && rs.Black[d.ID] {
			blackSkipped++
			continue
		}
		if rs != nil && rs.WhiteActive && !rs.White[d.ID] {
			whiteSkipped++
			continue
		}
		kept = append(kept, d)
	}
	return kept, whiteSkipped, blackSkipped
}

// RuleTargetDeviceIDs 将名单目标解析为设备 ID 列表
func RuleTargetDeviceIDs(targetType, targetID string) ([]string, error) {
	switch targetType {
	case types.RuleTargetDevice:
		if targetID == "" {
			return nil, fmt.Errorf("无效的设备ID")
		}
		return []string{targetID}, nil
	case types.RuleTargetGroup:
		groupID, err := strconv.ParseInt(targetID, 10, 64)
		if err != nil || groupID <= 0 {
			return nil, fmt.Errorf("无效的分组ID")
		}
		var devices []types.Device
		if err := db.Find("device", builder.Eq{"group_id": groupID}, &devices, 0, 0); err != nil {
			return nil, err
		}
		out := make([]string, 0, len(devices))
		for _, d := range devices {
			out = append(out, d.ID)
		}
		return out, nil
	case types.RuleTargetOrg:
		orgID, err := strconv.ParseInt(targetID, 10, 64)
		if err != nil || orgID <= 0 {
			return nil, fmt.Errorf("无效的组织ID")
		}
		orgIDs, err := orgSubtreeIDs(orgID)
		if err != nil {
			return nil, err
		}
		args, err := orgDeviceIDs(orgIDs)
		if err != nil {
			return nil, err
		}
		out := make([]string, 0, len(args))
		for _, id := range args {
			if s, ok := id.(string); ok {
				out = append(out, s)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("不支持的名单目标类型")
}
