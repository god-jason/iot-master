// 家庭成员管理接口
// 负责家庭成员的邀请、移除、角色更新，以及通过邀请码加入家庭等功能
package api

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/memstore"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// FamilyMemberHandler 家庭成员管理处理器
type FamilyMemberHandler struct {
	curd.Handler[types.Member]
}

// NewFamilyMemberHandler 创建家庭成员处理器实例（直接操作 member 表，ID 为 int64 自增）
func NewFamilyMemberHandler() *FamilyMemberHandler {
	return &FamilyMemberHandler{
		curd.Handler[types.Member]{
			Collection: "member",
			IsInt64ID:  true,
		},
	}
}

// Invite 邀请成员加入家庭（通过用户名）
func (h *FamilyMemberHandler) Invite(c *gin.Context) {
	var req struct {
		FamilyID int64  `json:"family_id"`
		Username string `json:"username"`
		Admin    bool   `json:"admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	if req.FamilyID == 0 {
		curd.Error(c, 400, "缺少家庭ID")
		return
	}
	if req.Username == "" {
		curd.Error(c, 400, "请输入用户名")
		return
	}

	// 权限校验：仅家庭所有者、家庭管理员或系统管理员可邀请成员
	user := GetCurrentUser(c)
	if !canManageFamily(user.UserID, req.FamilyID, user.Admin) {
		curd.Error(c, 403, "无权限邀请家庭成员")
		return
	}
	// 仅家庭所有者或系统管理员可以授予管理员角色，防止普通管理员提权
	if req.Admin && !user.Admin && !isFamilyOwner(user.UserID, req.FamilyID) {
		req.Admin = false
	}

	// 根据用户名查找目标用户
	var target types.User
	found, err := db.FindOne("user", builder.Eq{"username": req.Username}, &target)
	if err != nil || !found || target.ID == "" {
		curd.Error(c, 404, "用户不存在")
		return
	}
	if target.Disabled || target.Lock {
		curd.Error(c, 400, "目标用户已被禁用或锁定")
		return
	}

	// 检查该用户是否已经是该家庭的成员
	var existing types.Member
	found, err = db.FindOne("member", builder.Eq{"family_id": req.FamilyID, "user_id": target.ID}, &existing)
	if err == nil && found && existing.ID != 0 {
		curd.Error(c, 400, "该用户已是家庭成员")
		return
	}

	// 创建家庭成员记录（int64 自增 ID 由 MySQL 生成，不传 ID）
	member := types.Member{
		FamilyID: req.FamilyID,
		UserID:   target.ID,
		Admin:    req.Admin,
		Created:  time.Now(),
	}

	_, err = db.InsertOne("member", member)
	if err != nil {
		logger.Error("添加家庭成员失败", "error", err)
		curd.Error(c, 500, "邀请失败")
		return
	}

	curd.Success(c, member)
}

// Update 更新成员信息（目前仅支持管理员角色切换）
func (h *FamilyMemberHandler) Update(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的成员ID")
		return
	}

	var req struct {
		Admin bool `json:"admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 查询成员记录是否存在
	var member types.Member
	found, err := db.FindOne("member", builder.Eq{"id": id64}, &member)
	if err != nil || !found || member.ID == 0 {
		curd.Error(c, 404, "成员不存在")
		return
	}

	// 查询关联的家庭信息，用于权限校验
	var family types.Family
	found, err = db.FindOne("family", builder.Eq{"id": member.FamilyID}, &family)
	if err != nil || !found || family.ID == 0 {
		curd.Error(c, 404, "家庭不存在")
		return
	}

	// 权限校验：仅家庭所有者、家庭管理员或系统管理员可修改成员角色
	user := GetCurrentUser(c)
	if !canManageFamily(user.UserID, member.FamilyID, user.Admin) {
		curd.Error(c, 403, "权限不足")
		return
	}
	// 家庭管理员不能修改其他管理员（仅所有者/系统管理员可以）
	if member.Admin && member.UserID != user.UserID && !user.Admin && !isFamilyOwner(user.UserID, member.FamilyID) {
		curd.Error(c, 403, "无权限修改其他管理员")
		return
	}

	// 更新管理员角色
	_, err = db.UpdateOne("member", builder.Eq{"id": id64}, map[string]interface{}{"admin": req.Admin})
	if err != nil {
		logger.Error("更新成员信息失败", "error", err)
		curd.Error(c, 500, "更新失败")
		return
	}

	curd.Success(c, nil)
}

// Remove 移除家庭成员
func (h *FamilyMemberHandler) Remove(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的成员ID")
		return
	}

	// 查询成员记录
	var member types.Member
	found, err := db.FindOne("member", builder.Eq{"id": id64}, &member)
	if err != nil || !found || member.ID == 0 {
		curd.Error(c, 404, "成员不存在")
		return
	}

	// 查询家庭信息用于权限校验
	var family types.Family
	found, err = db.FindOne("family", builder.Eq{"id": member.FamilyID}, &family)
	if err != nil || !found || family.ID == 0 {
		curd.Error(c, 404, "家庭不存在")
		return
	}

	// 权限校验：成员可主动退出；否则需家庭所有者/管理员/系统管理员
	user := GetCurrentUser(c)
	if member.UserID != user.UserID && !canManageFamily(user.UserID, member.FamilyID, user.Admin) {
		curd.Error(c, 403, "权限不足")
		return
	}
	// 家庭管理员不能移除其他管理员（仅所有者/系统管理员可以）
	if member.Admin && member.UserID != user.UserID && !user.Admin && !isFamilyOwner(user.UserID, member.FamilyID) {
		curd.Error(c, 403, "无权限移除其他管理员")
		return
	}

	// 删除成员记录
	_, err = db.DeleteOne("member", builder.Eq{"id": id64})
	if err != nil {
		logger.Error("移除家庭成员失败", "error", err)
		curd.Error(c, 500, "移除失败")
		return
	}

	curd.Success(c, nil)
}

// AcceptInvite 接受家庭邀请（通过邀请码）
func (h *FamilyMemberHandler) AcceptInvite(c *gin.Context) {
	var req struct {
		InviteKey string `json:"invite_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	if req.InviteKey == "" {
		curd.Error(c, 400, "缺少邀请码")
		return
	}

	// 从缓存读取邀请码对应的数据
	data, err := memstore.Get("invite:" + req.InviteKey)
	if err != nil {
		curd.Error(c, 404, "邀请已过期或不存在")
		return
	}

	var inviteData map[string]string
	if err := json.Unmarshal([]byte(data), &inviteData); err != nil {
		curd.Error(c, 400, "邀请数据无效")
		return
	}

	familyIDStr := inviteData["family_id"]
	familyID, err := strconv.ParseInt(familyIDStr, 10, 64)
	if err != nil || familyID == 0 {
		curd.Error(c, 400, "邀请数据无效")
		return
	}

	// 邀请码必须对应一个仍然存在的家庭
	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": familyID}, &family)
	if err != nil || !found || family.ID == 0 {
		curd.Error(c, 404, "家庭不存在")
		return
	}

	// 获取当前登录用户ID
	userID := GetUserID(c)

	// 检查用户是否已是该家庭的成员
	var existing types.Member
	found, _ = db.FindOne("member", builder.Eq{"family_id": familyID, "user_id": userID}, &existing)
	if found && existing.ID != 0 {
		curd.Error(c, 400, "您已是该家庭的成员")
		return
	}

	// 创建新的家庭成员记录（int64 自增 ID 由 MySQL 生成）
	member := types.Member{
		FamilyID: familyID,
		UserID:   userID,
		Admin:    false,
		Created:  time.Now(),
	}

	_, err = db.InsertOne("member", member)
	if err != nil {
		logger.Error("添加家庭成员失败", "error", err)
		curd.Error(c, 500, "加入失败")
		return
	}

	// 删除已使用的邀请码
	memstore.Del("invite:" + req.InviteKey)

	curd.Success(c, gin.H{
		"family": family,
		"member": member,
	})
}
