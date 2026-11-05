// 登录与认证相关接口
// 提供用户名密码登录、扫码登录（会话管理+WebSocket实时通知）、二维码扫描识别等功能
// 扫码登录流程：生成会话 -> 前端展示二维码 -> 移动端扫码提交token -> WebSocket通知PC端登录结果
package api

import (
	"crypto/subtle"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/builder"
)

// loginSessionCookie 扫码登录会话绑定 Cookie 名（HttpOnly，防止仅凭会话 ID 窃取 token）
const loginSessionCookie = "dmp_login_session"

// loginSessionTTL 扫码登录会话有效期
const loginSessionTTL = 5 * time.Minute

// upgrader WebSocket升级器
// 将HTTP请求升级为WebSocket连接，用于扫码登录的实时通知
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // 读取缓冲区大小
	WriteBufferSize: 1024, // 写入缓冲区大小
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		return err == nil && strings.EqualFold(u.Host, r.Host)
	},
}

// ScanLoginHandler 扫码登录处理器
// 负责生成登录会话、WebSocket等待登录结果、提交登录token、二维码扫描识别等功能
type ScanLoginHandler struct{}

// NewScanLoginHandler 创建扫码登录处理器实例
func NewScanLoginHandler() *ScanLoginHandler {
	return &ScanLoginHandler{}
}

// LoginSession 登录会话结构体
// 用于存储扫码登录过程中的会话状态
type LoginSession struct {
	ID        string      // 会话ID
	Secret    string      // 会话密钥，仅通过 HttpOnly Cookie 下发，用于绑定发起扫码的浏览器
	Token     string      // 用户登录token（扫码成功后填充）
	ExpiresAt time.Time   // 过期时间
	Channel   chan string // 用于通知WebSocket客户端登录结果的通道
}

// loginSessions 登录会话映射表
// 使用sync.Map存储，支持并发访问，key为会话ID，value为LoginSession指针
var loginSessions sync.Map

// bindSessionCookie 将会话密钥写入 HttpOnly Cookie，绑定当前浏览器
func bindSessionCookie(c *gin.Context, session *LoginSession) {
	c.SetCookie(loginSessionCookie, session.ID+"."+session.Secret, int(loginSessionTTL.Seconds()), "/", "", false, true)
}

// currentSession 读取并校验会话：要求 URL 中的 id 与 Cookie 中的 id/secret 完全匹配。
// 仅知道会话 ID 无法读取 token，避免会话 ID 泄露导致登录态被窃取。
func currentSession(c *gin.Context, id string) (*LoginSession, bool) {
	if id == "" {
		return nil, false
	}
	v, exists := loginSessions.Load(id)
	if !exists {
		return nil, false
	}
	session, ok := v.(*LoginSession)
	if !ok || session == nil || time.Now().After(session.ExpiresAt) {
		loginSessions.Delete(id)
		return nil, false
	}
	cookie, err := c.Cookie(loginSessionCookie)
	if err != nil {
		return nil, false
	}
	parts := strings.SplitN(cookie, ".", 2)
	if len(parts) != 2 || parts[0] != id {
		return nil, false
	}
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(session.Secret)) != 1 {
		return nil, false
	}
	return session, true
}

// WaitLogin 生成或查询登录会话
// 路由: GET /login/wait
// 查询参数: id - 可选，会话ID
// 返回: 会话ID和登录URL（二维码内容）
// 说明: 传入 id 时仅当 Cookie 与会话匹配才返回 token；否则创建新会话
func (h *ScanLoginHandler) WaitLogin(c *gin.Context) {
	id := c.Query("id")

	// 如果传入了会话ID，检查是否存在且属于当前浏览器
	if id != "" {
		if session, ok := currentSession(c, id); ok {
			c.JSON(http.StatusOK, gin.H{
				"id":    id,
				"url":   fmt.Sprintf("noob-login:%s", id), // 二维码内容
				"token": session.Token,                    // 已扫码则返回token
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "会话不存在或已过期"})
		return
	}

	// 创建新的登录会话：ID 与密钥均为 32 位十六进制随机串
	sessionID := strings.ReplaceAll(uuid.NewString(), "-", "")
	session := &LoginSession{
		ID:        sessionID,
		Secret:    strings.ReplaceAll(uuid.NewString(), "-", ""),
		Token:     "",
		ExpiresAt: time.Now().Add(loginSessionTTL), // 5分钟过期
		Channel:   make(chan string, 1),            // 缓冲通道，用于接收登录token
	}
	loginSessions.Store(sessionID, session)
	bindSessionCookie(c, session)

	c.JSON(http.StatusOK, gin.H{
		"id":  sessionID,
		"url": fmt.Sprintf("noob-login:%s", sessionID), // 返回二维码内容
	})

	// 到期清理会话，不额外阻塞 goroutine
	time.AfterFunc(loginSessionTTL, func() { loginSessions.Delete(sessionID) })
}

// ScanWebsocket 通过WebSocket等待登录结果
// 路由: GET /login/ws
// 查询参数: id - 会话ID
// 返回: WebSocket连接，等待扫码结果
// 说明: 建立WebSocket连接后，等待5分钟内的扫码登录结果
//
//	成功时发送 {"code": 0, "token": "..."}
//	超时或失败时发送 {"code": 1, "message": "超时"}
func (h *ScanLoginHandler) ScanWebsocket(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少id参数"})
		return
	}

	// 查询会话（要求 Cookie 与会话匹配）
	session, ok := currentSession(c, id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "会话不存在或已过期"})
		return
	}

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 等待登录结果或超时
	select {
	case token := <-session.Channel: // 接收到登录token
		conn.WriteJSON(gin.H{"code": 0, "token": token})
	case <-time.After(loginSessionTTL): // 超时
		conn.WriteJSON(gin.H{"code": 1, "message": "超时"})
	}
}

// ScanResult 扫码结果结构体
// 用于提交扫码登录的token
type ScanResult struct {
	ID    string `json:"id"`    // 会话ID
	Token string `json:"token"` // 用户登录token
}

// SubmitToken 提交登录token
// 路由: POST /login/submit
// 请求体: ScanResult（id, token）
// 返回: 操作结果
// 说明: 客户端扫码后调用此接口提交用户token，服务器通知WebSocket等待端
func (h *ScanLoginHandler) SubmitToken(c *gin.Context) {
	var req ScanResult
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 校验提交的 token 必须是本平台签发的有效 JWT，拒绝伪造/过期 token
	claims, err := ParseToken(req.Token)
	if err != nil {
		curd.Error(c, 401, "无效或过期的token")
		return
	}
	// 扫码登录管理后台仅允许平台账号（普通用户账号只能在移动端使用）
	if !claims.Admin && !claims.Platform {
		curd.Error(c, 403, "普通用户账号仅支持移动端登录")
		return
	}
	// 扫码登录管理后台同样要求账号已关联组织（平台管理员除外）
	if !claims.Admin && !userHasOrgAssociation(claims.UserID) {
		curd.Error(c, 403, "当前账号未归属组织，无法登录管理后台")
		return
	}

	// 查询会话（要求 Cookie 与会话匹配，仅二维码发起方可以提交）
	session, ok := currentSession(c, req.ID)
	if !ok {
		curd.Error(c, 404, "会话不存在")
		return
	}

	// 设置token并通知WebSocket客户端（使用select避免通道已满时阻塞）
	session.Token = req.Token
	select {
	case session.Channel <- req.Token:
	default:
		// 通道已满，说明已处理过或等待端已断开
	}

	curd.Success(c, gin.H{"message": "登录成功"})
}

// LoginForm 登录表单结构体
type LoginForm struct {
	Username string `json:"username"`         // 用户名
	Password string `json:"password"`         // 密码
	Client   string `json:"client,omitempty"` // 登录端：admin（管理后台） / 空或 app（移动端）
}

// effectivePlatform 计算账号是否为平台账号：平台管理员一律视为平台账号
func effectivePlatform(user *types.User) bool {
	return user.Admin || user.Platform
}

// Login 用户名密码登录
// 路由: POST /login
// 请求体: LoginForm（username, password）
// 返回: 用户信息和登录token
// 说明: 验证用户名密码后生成JWT token，token有效期720小时（30天）
func (h *ScanLoginHandler) Login(c *gin.Context) {
	var req LoginForm
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}

	// 查询用户
	var user types.User
	found, err := db.FindOne("user", builder.Eq{"username": req.Username}, &user)
	if err != nil {
		logger.Error("查询用户失败", "error", err, "username", req.Username)
		curd.Error(c, 401, "用户名或密码错误")
		return
	}
	if !found || user.ID == "" {
		curd.Error(c, 401, "用户名或密码错误")
		return
	}

	// 禁用/锁定账号不允许登录
	if user.Disabled {
		curd.Error(c, 403, "账号已被禁用")
		return
	}
	if user.Lock {
		curd.Error(c, 403, "账号已被锁定")
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		curd.Error(c, 401, "用户名或密码错误")
		return
	}

	// 管理后台仅允许平台账号登录；普通用户账号只能通过移动端访问
	if req.Client == "admin" && !effectivePlatform(&user) {
		curd.Error(c, 403, "普通用户账号仅支持移动端登录")
		return
	}

	// 管理后台登录要求账号已关联组织（负责人 / 成员），平台管理员除外
	if req.Client == "admin" && !user.Admin && !userHasOrgAssociation(user.ID) {
		curd.Error(c, 403, "当前账号未归属组织，无法登录管理后台")
		return
	}

	// 生成JWT token
	platform := effectivePlatform(&user)
	token, err := GenerateToken(user.ID, 0, user.Admin, platform, TokenExpireHours)
	if err != nil {
		curd.Error(c, 500, "生成token失败")
		return
	}

	// 清空密码字段并回填平台账号标识
	user.Password = ""
	user.Platform = platform
	curd.Success(c, gin.H{"token": token, "user": user})
}
