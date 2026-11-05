// Package api 资源归属校验工具
// 统一提供家庭 / 设备 / 房间 / 场景 / 任务的访问权限判断，避免各接口重复实现导致越权
package api

import (
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"strconv"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// IsAdminUser 当前登录用户是否为系统管理员
func IsAdminUser(c *gin.Context) bool {
	return GetCurrentUser(c).Admin
}

// requireAdmin 要求系统管理员权限；无权限时写入 403 响应并返回 false
func requireAdmin(c *gin.Context) bool {
	if !IsAdminUser(c) {
		curd.Error(c, 403, "需要管理员权限")
		return false
	}
	return true
}

// canAccessFamily 判断用户是否有权访问家庭（所有者 / 家庭成员 / 系统管理员）
func canAccessFamily(userID string, familyID int64, admin bool) bool {
	if admin {
		return true
	}
	if userID == "" || familyID == 0 {
		return false
	}

	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": familyID}, &family)
	if err != nil || !found || family.ID == 0 {
		return false
	}
	if family.UserID == userID {
		return true
	}

	var member types.Member
	found, err = db.FindOne("member", builder.Eq{"family_id": familyID, "user_id": userID}, &member)
	return err == nil && found && member.ID != 0
}

// isFamilyOwner 判断用户是否为家庭所有者
func isFamilyOwner(userID string, familyID int64) bool {
	if userID == "" || familyID == 0 {
		return false
	}
	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": familyID}, &family)
	return err == nil && found && family.ID != 0 && family.UserID == userID
}

// canManageFamily 判断用户是否有权管理家庭（家庭所有者 / 家庭管理员 / 系统管理员）
func canManageFamily(userID string, familyID int64, admin bool) bool {
	if admin {
		return true
	}
	if userID == "" || familyID == 0 {
		return false
	}

	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": familyID}, &family)
	if err != nil || !found || family.ID == 0 {
		return false
	}
	if family.UserID == userID {
		return true
	}

	var member types.Member
	found, err = db.FindOne("member", builder.Eq{"family_id": familyID, "user_id": userID}, &member)
	return err == nil && found && member.ID != 0 && member.Admin
}

// canAccessRoom 判断用户是否有权访问房间（通过房间所属家庭判断）
func canAccessRoom(userID string, roomID int64, admin bool) bool {
	if admin {
		return true
	}
	if roomID == 0 {
		return false
	}
	var room types.Room
	found, err := db.FindOne("room", builder.Eq{"id": roomID}, &room)
	if err != nil || !found || room.ID == 0 {
		return false
	}
	return canAccessFamily(userID, room.FamilyID, false)
}

// canAccessDevice 判断用户是否有权访问设备（直接归属 / 家庭共享 / 组织绑定 / 系统管理员）
func canAccessDevice(c *gin.Context, device *types.Device) bool {
	user := GetCurrentUser(c)
	if user.Admin {
		return true
	}
	if device == nil || user.UserID == "" {
		return false
	}
	if device.UserID == user.UserID {
		return true
	}
	if device.FamilyID != 0 && canAccessFamily(user.UserID, device.FamilyID, false) {
		return true
	}
	// 组织负责人 / 成员可管理其组织子树内绑定的设备
	orgIDs := userOrgScopeIDs(c)
	return len(orgIDs) > 0 && deviceBoundToOrgs(device.ID, orgIDs)
}

// deviceBoundToOrgs 判断设备是否绑定到指定组织集合中的任意组织
func deviceBoundToOrgs(deviceID string, orgIDs []int64) bool {
	if deviceID == "" || len(orgIDs) == 0 {
		return false
	}
	n, err := db.Count("device_org", builder.And(
		builder.Eq{"device_id": deviceID},
		builder.In("org_id", int64ToInterfaces(orgIDs)...),
	))
	return err == nil && n > 0
}

// loadDevice 按 ID 读取设备
func loadDevice(id string) (*types.Device, bool) {
	if id == "" {
		return nil, false
	}
	var device types.Device
	found, err := db.FindOne("device", builder.Eq{"id": id}, &device)
	if err != nil || !found || device.ID == "" {
		return nil, false
	}
	return &device, true
}

// loadAccessibleDevice 读取设备并校验当前用户权限；失败时已写入 HTTP 响应
func loadAccessibleDevice(c *gin.Context, id string) (*types.Device, bool) {
	device, found := loadDevice(id)
	if !found {
		curd.Error(c, 404, "设备不存在")
		return nil, false
	}
	if !canAccessDevice(c, device) {
		curd.Error(c, 403, "无权限访问该设备")
		return nil, false
	}
	return device, true
}

// requireDeviceRefs 校验设备请求中携带的归属字段：普通用户只能归属自己可访问的家庭/房间/用户
func requireDeviceRefs(c *gin.Context, device *types.Device) bool {
	user := GetCurrentUser(c)
	if !user.Admin {
		if device.UserID != "" && device.UserID != user.UserID {
			curd.Error(c, 403, "无权限指定设备归属用户")
			return false
		}
		if device.FamilyID != 0 && !canAccessFamily(user.UserID, device.FamilyID, false) {
			curd.Error(c, 403, "无权限指定该家庭")
			return false
		}
		if device.RoomID != 0 && !canAccessRoom(user.UserID, device.RoomID, false) {
			curd.Error(c, 403, "无权限指定该房间")
			return false
		}
	}
	return true
}

// toInt64 将 JSON 反序列化后的数值安全转换为 int64
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
}

// requireFamilyScope 校验场景/任务的家庭归属：family_id 为 0 视为全局资源，仅管理员可操作
func requireFamilyScope(c *gin.Context, familyID int64) bool {
	user := GetCurrentUser(c)
	if familyID == 0 {
		if !user.Admin {
			curd.Error(c, 403, "仅管理员可操作全局资源")
			return false
		}
		return true
	}
	if !canAccessFamily(user.UserID, familyID, user.Admin) {
		curd.Error(c, 403, "无权限操作该家庭的资源")
		return false
	}
	return true
}

// loadAccessibleScene 读取场景并校验当前用户权限
func loadAccessibleScene(c *gin.Context, id int64) (*types.Scene, bool) {
	var scene types.Scene
	found, err := db.FindOne("scene", builder.Eq{"id": id}, &scene)
	if err != nil || !found || scene.ID == 0 {
		curd.Error(c, 404, "场景不存在")
		return nil, false
	}
	user := GetCurrentUser(c)
	// family_id 为 0 视为全局场景，仅管理员可操作；否则按家庭权限判断
	if scene.FamilyID == 0 {
		if !user.Admin {
			curd.Error(c, 403, "无权限操作该场景")
			return nil, false
		}
	} else if !canAccessFamily(user.UserID, scene.FamilyID, user.Admin) {
		curd.Error(c, 403, "无权限操作该场景")
		return nil, false
	}
	return &scene, true
}

// loadAccessibleTask 读取定时任务，并通过其关联场景校验权限
func loadAccessibleTask(c *gin.Context, id int64) (*types.Task, bool) {
	var task types.Task
	found, err := db.FindOne("task", builder.Eq{"id": id}, &task)
	if err != nil || !found || task.ID == 0 {
		curd.Error(c, 404, "任务不存在")
		return nil, false
	}
	var scene types.Scene
	found, err = db.FindOne("scene", builder.Eq{"id": task.SceneID}, &scene)
	if err != nil || !found || scene.ID == 0 {
		curd.Error(c, 404, "关联场景不存在")
		return nil, false
	}
	user := GetCurrentUser(c)
	if scene.FamilyID == 0 {
		if !user.Admin {
			curd.Error(c, 403, "无权限操作该任务")
			return nil, false
		}
	} else if !canAccessFamily(user.UserID, scene.FamilyID, user.Admin) {
		curd.Error(c, 403, "无权限操作该任务")
		return nil, false
	}
	return &task, true
}
