package ruleengine

import (
	"sync"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"

	"xorm.io/builder"
)

// scopeMatcher 规则作用域匹配器，负责判断设备是否落入规则范围，
// 以及定时触发时枚举范围内的全部设备。
type scopeMatcher struct {
	models ModelLookup // 设备 → 物模型映射查询

	groupCache   map[string]groupCacheEntry // 分组 ID → 分组设备集合（TTL 30 秒）
	groupCacheMu sync.RWMutex               // 保护 groupCache
}

// groupCacheEntry 分组设备集合的缓存项，expire 为过期时间（TTL 30 秒）。
type groupCacheEntry struct {
	devices map[string]struct{}
	expire  time.Time
}

// newScopeMatcher 构造作用域匹配器，models 用于按物模型判断设备归属。
func newScopeMatcher(models ModelLookup) *scopeMatcher {
	return &scopeMatcher{
		models:     models,
		groupCache: make(map[string]groupCacheEntry),
	}
}

// match 判断设备是否命中规则作用域。
// single 直接比对设备 ID；model 比较设备所属物模型；group 查分组设备集合（带缓存）。
func (m *scopeMatcher) match(scope types.RuleScope, deviceID string) bool {
	switch scope.Type {
	case types.RuleScopeSingle:
		return scope.DeviceID != "" && scope.DeviceID == deviceID
	case types.RuleScopeModel:
		return scope.ModelID != "" && m.models.ModelIDOf(deviceID) == scope.ModelID
	case types.RuleScopeGroup:
		if scope.GroupID == "" {
			return false
		}
		_, ok := m.groupDevices(scope.GroupID)[deviceID]
		return ok
	default:
		return false
	}
}

// groupDevices 查询分组下全部设备ID（带 30 秒缓存），数据来源 device.group_id。
// 先读缓存，未命中或已过期则回源查库并回填缓存；查库失败返回空集合（不缓存）。
func (m *scopeMatcher) groupDevices(groupID string) map[string]struct{} {
	m.groupCacheMu.RLock()
	if entry, ok := m.groupCache[groupID]; ok && time.Now().Before(entry.expire) {
		// 缓存命中且未过期，直接返回共享集合（调用方只读）
		m.groupCacheMu.RUnlock()
		return entry.devices
	}
	m.groupCacheMu.RUnlock()

	// 缓存缺失：回源查询的一次性设备集合
	set := make(map[string]struct{})
	var devices []types.Device
	if err := db.Find("device", builder.Eq{"group_id": groupID}, &devices, 0, 0); err == nil {
		for _, d := range devices {
			set[d.ID] = struct{}{}
		}
	}

	// 回填缓存（含查库失败时的空集合），避免短时间内重复打库
	m.groupCacheMu.Lock()
	m.groupCache[groupID] = groupCacheEntry{devices: set, expire: time.Now().Add(30 * time.Second)}
	m.groupCacheMu.Unlock()
	return set
}

// enumerate 枚举作用域内的全部设备ID（定时触发使用）。
// single 返回该设备；model 按 model_id 查库；group 复用分组设备缓存。
func (m *scopeMatcher) enumerate(scope types.RuleScope) []string {
	switch scope.Type {
	case types.RuleScopeSingle:
		if scope.DeviceID == "" {
			return nil
		}
		return []string{scope.DeviceID}
	case types.RuleScopeModel:
		if scope.ModelID == "" {
			return nil
		}
		var devices []types.Device
		if err := db.Find("device", builder.Eq{"model_id": scope.ModelID}, &devices, 0, 0); err != nil {
			return nil
		}
		ids := make([]string, 0, len(devices))
		for _, d := range devices {
			ids = append(ids, d.ID)
		}
		return ids
	case types.RuleScopeGroup:
		set := m.groupDevices(scope.GroupID)
		ids := make([]string, 0, len(set))
		for id := range set {
			ids = append(ids, id)
		}
		return ids
	default:
		return nil
	}
}
