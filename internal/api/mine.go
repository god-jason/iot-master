// Package api 个人中心接口处理模块
// 提供获取当前登录用户信息、更新个人资料、修改密码等个人账号相关接口
package api

import (
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/builder"
)

// MineHandler 个人中心处理器
type MineHandler struct{}

// NewMineHandler 创建个人中心处理器实例
func NewMineHandler() *MineHandler {
	return &MineHandler{}
}

// Me 获取当前登录用户信息
func (h *MineHandler) Me(c *gin.Context) {
	userID := GetCurrentUserID(c)

	// 根据 user_id 查询用户完整信息
	var u types.User
	found, err := db.FindOne("user", builder.Eq{"id": userID}, &u)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found || u.ID == "" {
		curd.Error(c, 404, "用户不存在")
		return
	}

	// 清空密码字段，避免返回敏感信息
	u.Password = ""
	curd.Success(c, u)
}

// UpdateProfile 更新当前登录用户的个人资料
func (h *MineHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname  string `json:"nickname"`
		Email     string `json:"email"`
		Cellphone string `json:"cellphone"`
		Avatar    string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	userID := GetCurrentUserID(c)

	// 按白名单字段更新
	update := map[string]interface{}{
		"nickname":  req.Nickname,
		"email":     req.Email,
		"cellphone": req.Cellphone,
		"avatar":    req.Avatar,
		"updated":   time.Now(),
	}
	if _, err := db.UpdateOne("user", builder.Eq{"id": userID}, update); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 回查更新后的完整用户信息返回给前端
	var u types.User
	found, err := db.FindOne("user", builder.Eq{"id": userID}, &u)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found || u.ID == "" {
		curd.Error(c, 404, "用户不存在")
		return
	}

	u.Password = ""
	curd.Success(c, u)
}

// ChangePassword 修改当前登录用户密码
func (h *MineHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 新密码不能为空
	if req.NewPassword == "" {
		curd.Error(c, 400, "新密码不能为空")
		return
	}

	userID := GetCurrentUserID(c)

	// 查询当前用户，获取已存储的密码哈希
	var u types.User
	found, err := db.FindOne("user", builder.Eq{"id": userID}, &u)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found || u.ID == "" {
		curd.Error(c, 404, "用户不存在")
		return
	}

	// 锁定账号禁止修改密码
	if u.Lock {
		curd.Error(c, 403, "账号已锁定，无法修改密码")
		return
	}

	// 校验原密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.OldPassword)); err != nil {
		curd.Error(c, 400, "原密码错误")
		return
	}

	// 使用 bcrypt 对新密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		curd.Error(c, 500, "密码加密失败")
		return
	}

	// 更新用户密码字段并刷新更新时间
	_, err = db.UpdateOne("user", builder.Eq{"id": userID}, map[string]interface{}{
		"password": string(hashedPassword),
		"updated":  time.Now(),
	})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	curd.Success(c, nil)
}
