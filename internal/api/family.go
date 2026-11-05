// 家庭管理接口
// 提供家庭的创建、查询、更新、删除、家庭列表查询，以及家庭邀请码生成功能
package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/memstore"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"xorm.io/builder"
)

// FamilyHandler 家庭管理处理器
// 负责家庭的增删改查、家庭列表与邀请码生成
type FamilyHandler struct{}

// NewFamilyHandler 创建家庭管理处理器实例
func NewFamilyHandler() *FamilyHandler {
	return &FamilyHandler{}
}

// pathID 解析路径参数为 int64（Family/FamilyID 都是 int64 自增）
func pathID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// Get 获取家庭详情
func (h *FamilyHandler) Get(c *gin.Context) {
	id64, err := pathID(c)
	if err != nil {
		curd.Error(c, 400, "无效的家庭ID")
		return
	}

	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": id64}, &family)
	if err != nil {
		logger.Error("查询家庭失败", "error", err)
		curd.Error(c, 500, "查询失败")
		return
	}
	if !found || family.ID == 0 {
		curd.Error(c, 404, "家庭不存在")
		return
	}

	// 权限校验：家庭所有者、家庭成员或系统管理员
	user := GetCurrentUser(c)
	if !canAccessFamily(user.UserID, id64, user.Admin) {
		curd.Error(c, 403, "无权限访问该家庭")
		return
	}

	curd.Success(c, family)
}

// Create 创建家庭
func (h *FamilyHandler) Create(c *gin.Context) {
	var family types.Family
	if err := c.ShouldBindJSON(&family); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	userID := GetCurrentUserID(c)
	family.UserID = userID
	family.Created = time.Now()

	id, err := db.InsertOne("family", &family)
	if err != nil {
		logger.Error("创建家庭失败", "error", err)
		curd.Error(c, 500, "创建失败")
		return
	}
	if id > 0 {
		family.ID = id
	}

	curd.Success(c, family)
}

// Update 更新家庭
func (h *FamilyHandler) Update(c *gin.Context) {
	id64, err := pathID(c)
	if err != nil {
		curd.Error(c, 400, "无效的家庭ID")
		return
	}

	var family types.Family
	if err := c.ShouldBindJSON(&family); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 权限校验：家庭所有者、家庭管理员或系统管理员
	user := GetCurrentUser(c)
	if !canManageFamily(user.UserID, id64, user.Admin) {
		curd.Error(c, 403, "无权限修改该家庭")
		return
	}

	updateSet := map[string]interface{}{}
	if family.Name != "" {
		updateSet["name"] = family.Name
	}
	if family.Icon != "" {
		updateSet["icon"] = family.Icon
	}

	if len(updateSet) == 0 {
		curd.Error(c, 400, "没有需要更新的字段")
		return
	}

	_, err = db.UpdateOne("family", builder.Eq{"id": id64}, updateSet)
	if err != nil {
		logger.Error("更新家庭失败", "error", err)
		curd.Error(c, 500, "更新失败")
		return
	}

	curd.Success(c, nil)
}

// Delete 删除家庭（级联删除成员、房间、绑定关系）
func (h *FamilyHandler) Delete(c *gin.Context) {
	id64, err := pathID(c)
	if err != nil {
		curd.Error(c, 400, "无效的家庭ID")
		return
	}

	userID := GetCurrentUserID(c)
	admin := IsAdminUser(c)

	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": id64}, &family)
	if err != nil {
		logger.Error("查询家庭失败", "error", err)
		curd.Error(c, 500, "查询失败")
		return
	}
	if !found || family.ID == 0 {
		curd.Error(c, 404, "家庭不存在")
		return
	}
	if family.UserID != userID && !admin {
		curd.Error(c, 403, "无权限删除")
		return
	}

	_, err = db.DeleteOne("family", builder.Eq{"id": id64})
	if err != nil {
		logger.Error("删除家庭失败", "error", err)
		curd.Error(c, 500, "删除失败")
		return
	}

	// 级联：清空该家庭下所有设备的归属字段（不再需要 binding 表，直接 UPDATE device）
	db.UpdateOne("device", builder.Eq{"family_id": id64}, map[string]interface{}{
		"family_id": nil,
		"room_id":   nil,
	})
	// 级联删除相关数据
	_, _ = db.DeleteMany("member", builder.Eq{"family_id": id64})
	_, _ = db.DeleteMany("room", builder.Eq{"family_id": id64})

	curd.Success(c, nil)
}

// List 获取我的家庭列表（用户拥有的 + 用户加入的）
func (h *FamilyHandler) List(c *gin.Context) {
	userID := GetCurrentUserID(c)

	// 查询用户加入的其他家庭（作为成员）
	var members []types.Member
	err := db.Find("member", builder.Eq{"user_id": userID}, &members, 0, 0)
	if err != nil {
		logger.Error("查询家庭成员关系失败", "error", err)
		curd.Error(c, 500, "查询失败")
		return
	}

	var cond builder.Cond
	if len(members) > 0 {
		familyIDs := make([]interface{}, 0, len(members))
		for _, m := range members {
			familyIDs = append(familyIDs, m.FamilyID)
		}
		cond = builder.Or(
			builder.Eq{"user_id": userID},
			builder.In("id", familyIDs...),
		)
	} else {
		cond = builder.Eq{"user_id": userID}
	}

	var families []types.Family
	err = db.Find("family", cond, &families, 0, 0)
	if err != nil {
		logger.Error("查询家庭列表失败", "error", err)
		curd.Error(c, 500, "查询失败")
		return
	}

	curd.Success(c, families)
}

// FamilyInviteResponse 家庭邀请码响应
// InviteKey 为缓存中的唯一邀请码，URL 为 App 端可识别的邀请串，ExpiresAt 为过期时间戳
type FamilyInviteResponse struct {
	InviteKey string `json:"invite_key"`
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at"`
}

// CreateInvite 创建家庭邀请码
func (h *FamilyHandler) CreateInvite(c *gin.Context) {
	id64, err := pathID(c)
	if err != nil {
		curd.Error(c, 400, "无效的家庭ID")
		return
	}

	userID := GetCurrentUserID(c)

	var family types.Family
	found, err := db.FindOne("family", builder.Eq{"id": id64}, &family)
	if err != nil || !found || family.ID == 0 {
		curd.Error(c, 404, "家庭不存在")
		return
	}

	// 权限校验：家庭所有者或家庭管理员
	if family.UserID != userID {
		var member types.Member
		found, err = db.FindOne("member", builder.Eq{"family_id": id64, "user_id": userID}, &member)
		if err != nil || !found || member.ID == 0 || !member.Admin {
			curd.Error(c, 403, "只有家庭所有者和管理员才能创建邀请")
			return
		}
	}

	roomID := c.Query("room_id")

	// 生成唯一邀请码
	inviteKey := uuid.New().String()

	// 构建邀请数据（family_id 转 string 存缓存）
	dataMap := map[string]string{
		"family_id": strconv.FormatInt(id64, 10),
	}
	if roomID != "" {
		dataMap["room_id"] = roomID
	}

	data, _ := json.Marshal(dataMap)

	// 存储邀请码到缓存，有效期5分钟
	memstore.Set("invite:"+inviteKey, string(data), 5*time.Minute)

	curd.Success(c, FamilyInviteResponse{
		InviteKey: inviteKey,
		URL:       fmt.Sprintf("noob-invite:%s", inviteKey),
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	})
}
