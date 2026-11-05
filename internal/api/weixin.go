// 微信相关接口
// 提供微信小程序登录、手机号获取、用户授权注册、头像上传等功能
// 对接微信开放平台API，实现微信生态用户体系与本系统的打通
package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/internal/weixin"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"xorm.io/builder"
)

// WeixinHandler 微信功能处理器
type WeixinHandler struct{}

// NewWeixinHandler 创建微信功能处理器实例
func NewWeixinHandler() *WeixinHandler {
	return &WeixinHandler{}
}

// WechatSessionRequest 微信登录凭证请求结构体
type WechatSessionRequest struct {
	Code string `json:"code" binding:"required"`
}

// Code2Session 微信小程序登录凭证校验
func (h *WeixinHandler) Code2Session(c *gin.Context) {
	var req WechatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	session, err := weixin.CodeToSession(req.Code)
	if err != nil {
		logger.Error("微信code2session失败", "error", err)
		curd.Error(c, 500, "微信授权失败")
		return
	}

	var user types.User
	var found bool

	// 优先使用UnionID查询
	if session.UnionID != "" {
		found, _ = db.FindOne("user", builder.Eq{"unionid": session.UnionID}, &user)
	}

	// UnionID未找到则降级使用OpenID查询
	if !found || user.ID == "" {
		found, _ = db.FindOne("user", builder.Eq{"openid": session.OpenID}, &user)
	}

	// 找到已注册用户
	if found && user.ID != "" {
		if user.Disabled || user.Lock {
			curd.Error(c, 403, "账号已被禁用或锁定")
			return
		}
		// 若当前用户记录的UnionID与最新返回不一致，自动补录更新
		if session.UnionID != "" && user.UnionID != session.UnionID {
			db.UpdateOne("user", builder.Eq{"id": user.ID}, map[string]interface{}{"unionid": session.UnionID})
			user.UnionID = session.UnionID
		}

		token, err := GenerateToken(user.ID, 0, user.Admin, effectivePlatform(&user), TokenExpireHours)
		if err != nil {
			curd.Error(c, 500, "生成token失败")
			return
		}
		user.Password = ""
		curd.Success(c, gin.H{"token": token, "user": user, "is_new": false})
		return
	}

	// 未找到对应用户，返回微信标识字段引导后续授权注册（不返回 session_key，避免会话密钥泄露）
	curd.Success(c, gin.H{
		"openid":    session.OpenID,
		"unionid":   session.UnionID,
		"need_auth": true,
	})
}

// WechatPhoneRequest 微信手机号解密请求结构体
type WechatPhoneRequest struct {
	Code string `json:"code" binding:"required"`
}

// GetPhoneNumber 通过微信小程序getPhoneNumber的code换取用户手机号
func (h *WeixinHandler) GetPhoneNumber(c *gin.Context) {
	var req WechatPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	phone, err := weixin.GetUserPhoneNumber(req.Code)
	if err != nil {
		logger.Error("获取微信手机号失败", "error", err)
		curd.Error(c, 500, "获取手机号失败")
		return
	}

	curd.Success(c, gin.H{"phone": phone})
}

// WechatAuthRequest 微信用户授权注册请求结构体
// 只接受微信登录 code，openid/unionid 一律由服务端换取，杜绝客户端伪造身份
type WechatAuthRequest struct {
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// Auth 微信用户授权注册或信息更新
func (h *WeixinHandler) Auth(c *gin.Context) {
	var req WechatAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 服务端用 code 换取 openid/unionid，禁止信任客户端提交的微信标识
	session, err := weixin.CodeToSession(req.Code)
	if err != nil {
		logger.Error("微信授权失败", "error", err)
		curd.Error(c, 500, "微信授权失败")
		return
	}
	if session.OpenID == "" {
		curd.Error(c, 400, "微信授权失败：openid 为空")
		return
	}

	var user types.User
	var found bool

	// 优先UnionID查询
	if session.UnionID != "" {
		found, _ = db.FindOne("user", builder.Eq{"unionid": session.UnionID}, &user)
	}
	// 降级OpenID查询
	if !found || user.ID == "" {
		found, _ = db.FindOne("user", builder.Eq{"openid": session.OpenID}, &user)
	}

	// 用户已存在：按需更新用户信息字段，返回登录态
	if found && user.ID != "" {
		if user.Disabled || user.Lock {
			curd.Error(c, 403, "账号已被禁用或锁定")
			return
		}
		updateFields := map[string]interface{}{}
		if session.UnionID != "" && user.UnionID != session.UnionID {
			updateFields["unionid"] = session.UnionID
			user.UnionID = session.UnionID
		}
		if req.Nickname != "" && user.Nickname != req.Nickname {
			updateFields["nickname"] = req.Nickname
			user.Nickname = req.Nickname
		}
		if req.Avatar != "" && user.Avatar != req.Avatar {
			updateFields["avatar"] = req.Avatar
			user.Avatar = req.Avatar
		}
		if len(updateFields) > 0 {
			db.UpdateOne("user", builder.Eq{"id": user.ID}, updateFields)
		}

		token, err := GenerateToken(user.ID, 0, user.Admin, effectivePlatform(&user), TokenExpireHours)
		if err != nil {
			curd.Error(c, 500, "生成token失败")
			return
		}
		user.Password = ""
		curd.Success(c, gin.H{"token": token, "user": user, "is_new": false})
		return
	}

	// 用户不存在：创建新用户
	if req.Nickname == "" {
		req.Nickname = "微信用户"
	}

	user = types.User{
		ID:       db.NewID(),
		Username: "wx_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16],
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		OpenID:   session.OpenID,
		UnionID:  session.UnionID,
		Created:  time.Now(),
	}
	if _, err := db.InsertOne("user", user); err != nil {
		logger.Error("创建微信用户失败", "error", err)
		curd.Error(c, 500, "创建用户失败")
		return
	}

	token, err := GenerateToken(user.ID, 0, user.Admin, effectivePlatform(&user), TokenExpireHours)
	if err != nil {
		curd.Error(c, 500, "生成token失败")
		return
	}

	user.Password = ""
	curd.Success(c, gin.H{"token": token, "user": user, "is_new": true})
}

// UploadAvatar 上传微信头像（无需登录）
func (h *WeixinHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		curd.Error(c, 400, "请选择要上传的文件")
		return
	}

	if file.Size > 5*1024*1024 {
		curd.Error(c, 400, "文件大小不能超过5MB")
		return
	}

	uploadDir := config.GetString("upload", "dir")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// 仅允许图片扩展名，防止上传 html/svg 等可执行内容
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		curd.Error(c, 400, "仅支持 jpg/jpeg/png/webp 图片")
		return
	}

	filename := fmt.Sprintf("wx_avatar_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	relativePath := fmt.Sprintf("/images/%s", filename)
	fullPath := filepath.Join(uploadDir, "images", filename)

	if err := os.MkdirAll(filepath.Join(uploadDir, "images"), 0755); err != nil {
		curd.Error(c, 500, "无法创建上传目录")
		return
	}
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		curd.Error(c, 500, "无法保存文件")
		return
	}

	serverURL := config.GetString("upload", "server_url")
	if serverURL == "" {
		proto := "http"
		if c.Request.TLS != nil {
			proto = "https"
		}
		serverURL = fmt.Sprintf("%s://%s", proto, c.Request.Host)
	}
	serverURL = serverURL + relativePath

	curd.Success(c, gin.H{"url": serverURL})
}
