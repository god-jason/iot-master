// Package api 主 API 路由注册模块
// 初始化并注册所有业务路由分组，包括登录、微信、个人中心、设备、家庭、规则引擎、场景等接口
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/protocol"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/builder"
)

// InitRouter 初始化并注册主 API 路由
// 路由前缀由调用方通过 RouterGroup 传入（通常为 /api）
// 参数: r - Gin 路由分组实例，作为所有业务路由的父级
// 说明:
//   - 公开接口（OEM、扫码登录、微信）注册在 JWTAuth 之前
//   - 业务接口统一在 JWTAuth 之后注册，按领域拆分到各 registerXxxRoutes 函数
func InitRouter(r *gin.RouterGroup) {
	registerPublicRoutes(r)

	// 认证边界：其后注册的路由全部需要 JWT
	r.Use(JWTAuth())

	registerMineRoutes(r)
	registerAccountRoutes(r)
	registerOrgRoutes(r)
	r.GET("/user/search", SearchUsers) // 用户检索（负责人/成员选择器）
	registerDeviceRoutes(r)
	registerGroupRoutes(r)
	registerFamilyRoutes(r)
	registerRoomRoutes(r)
	registerMemberRoutes(r)
	registerDeviceLogRoutes(r)
	registerDeviceEventRoutes(r)
	registerRuleRoutes(r)
	registerSceneRoutes(r)
	registerTaskRoutes(r)
	registerModelRoutes(r)
	registerProtocolRoutes(r)
	registerScadaRoutes(r)
	registerComponentRoutes(r)
	registerUploadRoutes(r)
	registerGatewayRoutes(r)
	registerOTARoutes(r)
	registerMaintenanceRoutes(r)
}

// registerPublicRoutes 注册无需认证的公开路由（OEM 配置、扫码登录、微信小程序）
func registerPublicRoutes(r *gin.RouterGroup) {
	r.GET("/oem", GetOEM) // 获取品牌白标配置（logo、系统名称等）

	loginHandler := NewScanLoginHandler()
	login := r.Group("/login")
	{
		login.POST("", loginHandler.Login)              // 发起扫码登录，获取二维码
		login.GET("/wait", loginHandler.WaitLogin)      // 轮询等待扫码确认
		login.GET("/ws", loginHandler.ScanWebsocket)    // 扫码状态 WebSocket 推送
		login.POST("/submit", loginHandler.SubmitToken) // 扫码端提交确认 token
	}

	weixinHandler := NewWeixinHandler()
	weixinAPI := r.Group("/weixin")
	{
		weixinAPI.POST("/code2session", weixinHandler.Code2Session) // 微信 code 换 session
		weixinAPI.POST("/auth", weixinHandler.Auth)                 // 微信授权登录
		weixinAPI.POST("/avatar", weixinHandler.UploadAvatar)       // 上传微信头像
		weixinAPI.POST("/phone", weixinHandler.GetPhoneNumber)      // 获取微信手机号
	}

	// 固件下载：设备侧直接 HTTP 拉取，不能带 token；支持 Range 断点续传
	r.GET("/ota/firmware/:id/file", DownloadFirmware)
}

// registerMineRoutes 注册个人中心路由
func registerMineRoutes(r *gin.RouterGroup) {
	mineHandler := NewMineHandler()
	r.GET("/me", mineHandler.Me)                    // 获取当前登录用户信息
	r.PUT("/me", mineHandler.UpdateProfile)         // 更新当前登录用户个人资料
	r.POST("/password", mineHandler.ChangePassword) // 修改当前用户密码
}

// registerAccountRoutes 注册子账号管理路由
// 创建走自定义逻辑：只接收白名单字段，密码 bcrypt 加密，强制 Admin=false 且归属当前用户
func registerAccountRoutes(r *gin.RouterGroup) {
	accountHandler := NewAccountHandler()
	accounts := r.Group("/account")
	{
		accounts.POST("/search", accountHandler.Search) // 分页搜索子账号列表
		accounts.GET("/:id", accountHandler.Get)        // 获取单个子账号详情
		accounts.POST("", accountHandler.Create)        // 创建新子账号
		accounts.PUT("/:id", accountHandler.Update)     // 更新子账号信息
		accounts.DELETE("/:id", accountHandler.Delete)  // 删除子账号
	}
}

// registerOrgRoutes 注册组织管理路由（多级组织树，管理员管理全部，其他用户管理自身子树）
func registerOrgRoutes(r *gin.RouterGroup) {
	orgHandler := NewOrgHandler()
	orgs := r.Group("/org")
	{
		orgs.GET("/tree", orgHandler.Tree)                            // 组织树（按权限返回子树）
		orgs.GET("/:id", orgHandler.Get)                              // 组织详情
		orgs.POST("", orgHandler.Create)                              // 创建组织
		orgs.PUT("/:id", orgHandler.Update)                           // 更新组织
		orgs.DELETE("/:id", orgHandler.Delete)                        // 删除组织
		orgs.GET("/:id/members", orgHandler.ListMembers)              // 组织成员列表
		orgs.POST("/:id/members", orgHandler.AddMember)               // 添加成员
		orgs.DELETE("/:id/members/:user_id", orgHandler.RemoveMember) // 移除成员
	}
}

// registerDeviceRoutes 注册设备与子设备管理路由
func registerDeviceRoutes(r *gin.RouterGroup) {
	deviceHandler := NewDeviceHandler()
	subDeviceHandler := NewSubDeviceHandler()
	devices := r.Group("/device")
	{
		// 设备基础 CRUD
		devices.POST("/search", deviceHandler.Search) // 分页搜索设备列表
		devices.GET("/:id", deviceHandler.Get)        // 获取单个设备详情
		devices.POST("", deviceHandler.Create)        // 创建设备
		devices.PUT("/:id", deviceHandler.Update)     // 更新设备信息
		devices.DELETE("/:id", deviceHandler.Delete)  // 删除设备

		// 设备归属管理
		devices.PUT("/:id/unbind", deviceHandler.Unbind) // 清空设备归属（家庭/房间/分组）

		// 设备实时属性 / 历史数据 / 控制 / 配置下发
		devices.GET("/:id/property", deviceHandler.Property)    // 实时属性快照
		devices.GET("/:id/history", deviceHandler.History)      // 属性历史曲线
		devices.POST("/:id/control", deviceHandler.Control)     // 调用设备服务
		devices.POST("/:id/setting/set", deviceHandler.Setting) // 下发设备配置

		// 设备-组织绑定（一台设备可绑定多个组织）
		deviceOrgHandler := NewDeviceOrgHandler()
		devices.GET("/:id/orgs", deviceOrgHandler.List)              // 查询设备已绑定组织
		devices.POST("/:id/orgs", deviceOrgHandler.Bind)             // 绑定设备到组织（含分配至子级组织）
		devices.DELETE("/:id/orgs/:org_id", deviceOrgHandler.Unbind) // 解除设备与组织绑定

		// 子设备管理接口：网关下挂载的子设备 CRUD
		devices.POST("/:id/sub-device/search", subDeviceHandler.Search)           // 搜索子设备
		devices.GET("/:id/sub-device/:sub_device_id", subDeviceHandler.Get)       // 获取子设备详情
		devices.POST("/:id/sub-device", subDeviceHandler.Create)                  // 创建子设备
		devices.PUT("/:id/sub-device/:sub_device_id", subDeviceHandler.Update)    // 更新子设备
		devices.DELETE("/:id/sub-device/:sub_device_id", subDeviceHandler.Delete) // 删除子设备
	}
}

// registerGroupRoutes 注册设备分组管理路由
func registerGroupRoutes(r *gin.RouterGroup) {
	groupHandler := curd.Handler[types.Group]{
		Collection:   "group",
		IsInt64ID:    true, // int64 自增主键
		Keywords:     []string{"name"},
		UpdateFields: []string{"name", "parent_id"},
	}
	groups := r.Group("/group")
	{
		groups.POST("/search", groupHandler.Search) // 分页搜索分组列表
		groups.GET("/:id", groupHandler.Get)        // 获取单个分组详情
		groups.POST("", groupHandler.Create)        // 创建分组
		groups.PUT("/:id", groupHandler.Update)     // 更新分组
		groups.DELETE("/:id", groupHandler.Delete)  // 删除分组
	}
}

// registerFamilyRoutes 注册家庭管理路由
func registerFamilyRoutes(r *gin.RouterGroup) {
	familyHandler := NewFamilyHandler()
	families := r.Group("/family")
	{
		families.GET("", familyHandler.List)                    // 获取当前用户的家庭列表
		families.GET("/:id", familyHandler.Get)                 // 获取家庭详情
		families.POST("", familyHandler.Create)                 // 创建家庭
		families.PUT("/:id", familyHandler.Update)              // 更新家庭信息
		families.DELETE("/:id", familyHandler.Delete)           // 删除家庭
		families.GET("/:id/invite", familyHandler.CreateInvite) // 生成家庭邀请链接
	}
}

// registerRoomRoutes 注册房间管理路由
func registerRoomRoutes(r *gin.RouterGroup) {
	roomHandler := curd.Handler[types.Room]{
		Collection:   "room",
		IsInt64ID:    true, // int64 自增主键
		Keywords:     []string{"name"},
		UpdateFields: []string{"name", "family_id", "sort"},
	}
	rooms := r.Group("/room")
	{
		rooms.POST("/search", roomHandler.Search) // 分页搜索房间列表
		rooms.POST("", roomHandler.Create)        // 创建房间
		rooms.PUT("/:id", roomHandler.Update)     // 更新房间
		rooms.DELETE("/:id", roomHandler.Delete)  // 删除房间
	}
}

// registerMemberRoutes 注册家庭成员管理路由
func registerMemberRoutes(r *gin.RouterGroup) {
	memberHandler := NewFamilyMemberHandler()
	familyMembers := r.Group("/member")
	{
		familyMembers.POST("/search", memberHandler.Search)       // 搜索家庭成员列表
		familyMembers.POST("/invite", memberHandler.Invite)       // 邀请用户加入家庭
		familyMembers.POST("/accept", memberHandler.AcceptInvite) // 接受家庭邀请
		familyMembers.PUT("/:id", memberHandler.Update)           // 更新成员信息/角色
		familyMembers.DELETE("/:id", memberHandler.Remove)        // 移除家庭成员
	}
}

// registerDeviceLogRoutes 注册设备日志查询路由
func registerDeviceLogRoutes(r *gin.RouterGroup) {
	deviceLogHandler := curd.Handler[types.DeviceLog]{
		Collection:   "device_log",
		IsInt64ID:    true, // int64 自增主键
		Keywords:     []string{"device_id", "message"},
		UpdateFields: []string{"message", "level"},
	}
	deviceLogs := r.Group("/device-log")
	{
		deviceLogs.POST("/search", deviceLogHandler.Search) // 分页搜索设备日志
		deviceLogs.GET("/:id", deviceLogHandler.Get)        // 获取单条日志详情
	}
}

// registerDeviceEventRoutes 注册设备事件告警路由
func registerDeviceEventRoutes(r *gin.RouterGroup) {
	deviceEventHandler := NewDeviceEventHandler()
	deviceEvents := r.Group("/device-event")
	{
		deviceEvents.POST("/search", deviceEventHandler.Search)    // 分页搜索设备事件
		deviceEvents.GET("/:id", deviceEventHandler.Get)           // 获取事件详情
		deviceEvents.PUT("/:id", deviceEventHandler.Update)        // 更新事件信息
		deviceEvents.DELETE("/:id", deviceEventHandler.Delete)     // 删除事件
		deviceEvents.PUT("/:id/handle", deviceEventHandler.Handle) // 标记事件为已处理
	}
}

// registerRuleRoutes 注册规则引擎与规则日志路由
func registerRuleRoutes(r *gin.RouterGroup) {
	ruleHandler := NewRuleHandler()
	ruleLogHandler := NewRuleLogHandler()
	rules := r.Group("/rule")
	{
		// 注意：静态路径 /logs/search 需在通配路径之前语义清晰，避免与 /:id/test 混淆
		rules.POST("/logs/search", ruleLogHandler.Search) // 分页搜索规则执行日志
		rules.POST("/search", ruleHandler.Search)         // 分页搜索规则
		rules.GET("/:id", ruleHandler.Get)                // 获取规则详情
		rules.POST("", ruleHandler.Create)                // 创建规则
		rules.PUT("/:id", ruleHandler.Update)             // 更新规则
		rules.DELETE("/:id", ruleHandler.Delete)          // 删除规则
		rules.PUT("/:id/switch", ruleHandler.Switch)      // 启用/停用规则
		rules.POST("/:id/test", ruleHandler.Test)         // 模拟数据试跑规则
	}
}

// registerSceneRoutes 注册场景管理路由
func registerSceneRoutes(r *gin.RouterGroup) {
	sceneHandler := NewSceneHandler()
	scenes := r.Group("/scene")
	{
		scenes.POST("/search", sceneHandler.Search)       // 分页搜索场景
		scenes.GET("/:id", sceneHandler.Get)              // 获取场景详情
		scenes.POST("", sceneHandler.Create)              // 创建场景
		scenes.PUT("/:id", sceneHandler.Update)           // 更新场景
		scenes.DELETE("/:id", sceneHandler.Delete)        // 删除场景
		scenes.PUT("/:id/switch", sceneHandler.Switch)    // 启用/停用场景
		scenes.POST("/:id/execute", sceneHandler.Execute) // 手动执行场景
	}
}

// registerTaskRoutes 注册定时任务路由
func registerTaskRoutes(r *gin.RouterGroup) {
	taskHandler := NewTaskHandler()
	tasks := r.Group("/task")
	{
		tasks.POST("/search", taskHandler.Search)    // 分页搜索定时任务
		tasks.GET("/:id", taskHandler.Get)           // 获取任务详情
		tasks.POST("", taskHandler.Create)           // 创建定时任务
		tasks.PUT("/:id", taskHandler.Update)        // 更新定时任务
		tasks.DELETE("/:id", taskHandler.Delete)     // 删除定时任务
		tasks.PUT("/:id/switch", taskHandler.Switch) // 启用/停用任务
	}
}

// registerModelRoutes 注册物模型管理路由
func registerModelRoutes(r *gin.RouterGroup) {
	modelHandler := curd.Handler[types.Model]{
		Collection:   "model",                                                                                                                                       // MongoDB 集合名：model
		Keywords:     []string{"name"},                                                                                                                              // 模糊搜索字段：物模型名称
		UpdateFields: []string{"name", "icon", "type", "protocol", "properties", "pollers", "expressions", "filters", "events", "services", "settings", "controls"}, // 允许更新的完整字段
	}
	models := r.Group("/model")
	{
		models.POST("/search", modelHandler.Search) // 分页搜索物模型
		models.GET("/:id", modelHandler.Get)        // 获取物模型详情
		models.POST("", modelHandler.Create)        // 创建物模型
		models.PUT("/:id", updateModel)             // 更新物模型（支持修改物模型ID）
		models.DELETE("/:id", modelHandler.Delete)  // 删除物模型
	}
}

// registerProtocolRoutes 注册协议列表路由（协议由 pkg/protocol 代码定义，非数据库）
func registerProtocolRoutes(r *gin.RouterGroup) {
	protocols := r.Group("/protocol")
	{
		protocols.GET("", protocolList)         // 获取全部协议
		protocols.GET("/:name", protocolDetail) // 获取单个协议描述
	}
}

// protocolList 返回全部协议清单
func protocolList(c *gin.Context) {
	curd.Success(c, protocol.Protocols())
}

// protocolDetail 返回指定协议描述，不存在时 404
func protocolDetail(c *gin.Context) {
	desc := protocol.GetManifest(c.Param("name"))
	if desc.Name == "" {
		curd.Error(c, 404, "协议不存在")
		return
	}
	curd.Success(c, desc)
}

// registerScadaRoutes 注册组态/大屏画面管理路由
func registerScadaRoutes(r *gin.RouterGroup) {
	scadaHandler := curd.Handler[types.Scada]{
		Collection:   "scada",
		IsInt64ID:    true, // int64 自增主键
		Keywords:     []string{"name"},
		UpdateFields: []string{"name", "description", "icon", "type", "canvas", "elements", "disabled", "updated"},
	}
	scadaGroup := r.Group("/scada")
	{
		scadaGroup.POST("/search", scadaHandler.Search) // 分页搜索画面列表
		scadaGroup.GET("/:id", scadaHandler.Get)        // 获取画面详情
		scadaGroup.POST("", scadaHandler.Create)        // 创建画面
		scadaGroup.PUT("/:id", scadaHandler.Update)     // 更新画面
		scadaGroup.DELETE("/:id", scadaHandler.Delete)  // 删除画面
	}
}

// registerComponentRoutes 注册组态组件库路由
func registerComponentRoutes(r *gin.RouterGroup) {
	componentHandler := curd.Handler[types.Component]{
		Collection:   "component",
		Keywords:     []string{"name"},
		UpdateFields: []string{"name", "description", "icon", "category", "elements", "variables", "events", "scripts", "disabled", "updated"},
	}
	components := r.Group("/component")
	{
		components.POST("/search", componentHandler.Search) // 分页搜索组件列表
		components.GET("/:id", componentHandler.Get)        // 获取组件详情
		components.POST("", componentHandler.Create)        // 创建组件
		components.PUT("/:id", componentHandler.Update)     // 更新组件
		components.DELETE("/:id", componentHandler.Delete)  // 删除组件
	}
}

// registerUploadRoutes 注册文件上传路由
func registerUploadRoutes(r *gin.RouterGroup) {
	uploadHandler := NewUploadHandler()
	uploads := r.Group("/upload")
	{
		uploads.POST("/image", uploadHandler.UploadImage)           // 上传普通图片
		uploads.POST("/avatar", uploadHandler.UploadAvatar)         // 上传用户头像
		uploads.POST("/icon", uploadHandler.UploadIcon)             // 上传图标
		uploads.POST("/attachment", uploadHandler.UploadAttachment) // 上传附件文件
		uploads.POST("/raw", uploadHandler.UploadRaw)               // 上传原始二进制文件
		uploads.DELETE("", uploadHandler.DeleteFile)                // 删除已上传文件
	}
}

// registerOTARoutes 注册 OTA 固件与升级任务路由
func registerOTARoutes(r *gin.RouterGroup) {
	otaHandler := NewOTAHandler()

	// 固件上传与文件管理
	r.POST("/upload/firmware", otaHandler.UploadFirmware)
	firmware := r.Group("/firmware")
	{
		firmware.POST("/search", otaHandler.SearchFirmware)                   // 分页搜索固件
		firmware.GET("/:id", otaHandler.GetFirmware)                          // 固件详情
		firmware.PUT("/:id", otaHandler.UpdateFirmware)                       // 更新固件元信息
		firmware.DELETE("/:id", otaHandler.DeleteFirmware)                    // 删除固件
		firmware.PUT("/:id/release", otaHandler.ReleaseFirmware)              // 发布固件
		firmware.PUT("/:id/disable", otaHandler.DisableFirmware)              // 停用固件
		firmware.GET("/:id/rules", otaHandler.ListFirmwareRules)              // 黑白名单规则
		firmware.POST("/:id/rules", otaHandler.AddFirmwareRules)              // 添加黑白名单规则
		firmware.DELETE("/:id/rules/:rule_id", otaHandler.DeleteFirmwareRule) // 删除名单规则
	}

	// 升级任务
	tasks := r.Group("/ota/task")
	{
		tasks.POST("/preview", otaHandler.PreviewTask)         // 预览目标设备与批次
		tasks.POST("/search", otaHandler.SearchTasks)          // 分页搜索任务
		tasks.POST("", otaHandler.CreateTask)                  // 创建任务
		tasks.GET("/:id", otaHandler.GetTask)                  // 任务详情
		tasks.GET("/:id/devices", otaHandler.ListTaskDevices)  // 任务设备明细
		tasks.PUT("/:id/pause", otaHandler.PauseTask)          // 暂停
		tasks.PUT("/:id/resume", otaHandler.ResumeTask)        // 恢复
		tasks.PUT("/:id/cancel", otaHandler.CancelTask)        // 取消
		tasks.PUT("/:id/retry-failed", otaHandler.RetryFailed) // 重试失败设备
	}

	// 单设备升级（复用任务模型）
	r.POST("/device/:id/ota", otaHandler.UpgradeDevice)
}

// registerMaintenanceRoutes 注册维保管理路由
func registerMaintenanceRoutes(r *gin.RouterGroup) {
	planHandler := NewPlanHandler()
	orderHandler := NewOrderHandler()

	m := r.Group("/maintenance")
	{
		// 维保计划
		m.POST("/plan/search", planHandler.Search)         // 分页搜索计划
		m.GET("/plan/:id", planHandler.Get)                // 计划详情
		m.POST("/plan", planHandler.Create)                // 创建计划
		m.PUT("/plan/:id", planHandler.Update)             // 更新计划
		m.DELETE("/plan/:id", planHandler.Delete)          // 删除计划
		m.PUT("/plan/:id/switch", planHandler.Switch)      // 启用/停用计划
		m.POST("/plan/:id/generate", planHandler.Generate) // 立即生成工单

		// 维保工单：导出路由需在 /order/:id 之前注册
		m.GET("/order/export", orderHandler.Export)         // 导出工单 CSV
		m.POST("/order/search", orderHandler.Search)        // 分页搜索工单
		m.GET("/order/:id", orderHandler.Get)               // 工单详情
		m.POST("/order", orderHandler.Create)               // 手工创建工单
		m.PUT("/order/:id", orderHandler.Update)            // 更新工单
		m.DELETE("/order/:id", orderHandler.Delete)         // 删除工单
		m.PUT("/order/:id/assign", orderHandler.Assign)     // 指派负责人
		m.PUT("/order/:id/start", orderHandler.StartOrder)  // 开始处理
		m.PUT("/order/:id/complete", orderHandler.Complete) // 完成工单
		m.PUT("/order/:id/cancel", orderHandler.Cancel)     // 取消工单

		m.GET("/stats/summary", orderHandler.Summary) // 维保统计汇总
	}
}

// AccountUser 账号信息（附带所属组织，组织关系保存在 org_member）
type AccountUser struct {
	types.User
	OrgIDs []int64 `json:"org_ids,omitempty"`
}

// userOrgIDs 批量查询用户所属组织（组织关系来自 org_member）
func userOrgIDs(userIDs []string) map[string][]int64 {
	out := make(map[string][]int64, len(userIDs))
	args := make([]interface{}, 0, len(userIDs))
	for _, id := range userIDs {
		if id != "" {
			args = append(args, id)
		}
	}
	if len(args) == 0 {
		return out
	}
	var rels []types.OrgMember
	if err := db.Find("org_member", builder.In("user_id", args...), &rels, 0, 0); err != nil {
		return out
	}
	for _, r := range rels {
		out[r.UserID] = append(out[r.UserID], r.OrgID)
	}
	return out
}

// syncUserOrg 同步账号所属组织：先移除该用户在其它组织中的成员关系，再确保目标组织成员关系存在
func syncUserOrg(userID string, orgID int64) error {
	if _, err := db.DeleteOne("org_member", builder.And(builder.Eq{"user_id": userID}, builder.Neq{"org_id": orgID})); err != nil {
		return err
	}
	if orgID <= 0 {
		return nil
	}
	if n, _ := db.Count("org_member", builder.Eq{"user_id": userID, "org_id": orgID}); n > 0 {
		return nil
	}
	rel := types.OrgMember{OrgID: orgID, UserID: userID, Created: time.Now()}
	_, err := db.InsertOne("org_member", &rel)
	return err
}

// AccountHandler 子账号管理处理器
// 在通用 curd.Handler 基础上重写 Create：只接收白名单字段、密码加密、禁止提权
type AccountHandler struct {
	curd.Handler[types.User]
}

// NewAccountHandler 创建子账号处理器：默认只查询/操作当前用户创建的子账号
func NewAccountHandler() *AccountHandler {
	return &AccountHandler{
		Handler: curd.Handler[types.User]{
			Collection:   "user",
			Keywords:     []string{"nickname"},
			Filter:       map[string]string{"parent_id": "user_id"},
			UpdateFields: []string{"nickname", "avatar", "email", "cellphone", "disabled", "platform"},
		},
	}
}

// Create 创建子账号：字段白名单 + bcrypt 加密密码，强制非管理员并归属当前用户
func (h *AccountHandler) Create(c *gin.Context) {
	var req struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Email     string `json:"email"`
		Cellphone string `json:"cellphone"`
		OrgID     int64  `json:"org_id"`
		Platform  bool   `json:"platform"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	if req.Username == "" || req.Password == "" {
		curd.Error(c, 400, "用户名和密码不能为空")
		return
	}
	if !requireOrgAssignable(c, req.OrgID) {
		return
	}

	// 用户名唯一性校验
	var existing types.User
	if found, _ := db.FindOne("user", builder.Eq{"username": req.Username}, &existing); found {
		curd.Error(c, 400, "用户名已存在")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		curd.Error(c, 500, "密码加密失败")
		return
	}

	user := types.User{
		ID:        db.NewID(),
		Username:  req.Username,
		Password:  string(hashed),
		Nickname:  req.Nickname,
		Avatar:    req.Avatar,
		Email:     req.Email,
		Cellphone: req.Cellphone,
		ParentID:  GetCurrentUserID(c),
		Platform:  req.Platform,
		Created:   time.Now(),
	}
	if _, err := db.InsertOne("user", user); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 所属组织统一记录在 org_member 成员关系中
	if req.OrgID > 0 {
		rel := types.OrgMember{OrgID: req.OrgID, UserID: user.ID, Created: time.Now()}
		if _, err := db.InsertOne("org_member", &rel); err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
	}

	user.Password = "" // 不返回密码哈希
	curd.Success(c, AccountUser{User: user, OrgIDs: userOrgIDs([]string{user.ID})[user.ID]})
}

// Update 更新子账号：若变更所属组织，需校验目标组织在自身权限范围内
func (h *AccountHandler) Update(c *gin.Context) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	// 所属组织不再写入用户表，改为同步 org_member 成员关系
	var orgID int64
	orgProvided := false
	if v, ok := payload["org_id"]; ok {
		orgID = toInt64(v)
		if !requireOrgAssignable(c, orgID) {
			return
		}
		orgProvided = true
		delete(payload, "org_id")
	}
	// 平台账号标识必须是布尔值
	if v, ok := payload["platform"]; ok {
		if _, isBool := v.(bool); !isBool {
			curd.Error(c, 400, "无效的平台账号标识")
			return
		}
	}

	// 账号必须存在且在当前用户管辖范围内
	var user types.User
	found, err := db.FindOne("user", builder.And(builder.Eq{"id": c.Param("id")}, builder.Eq{"parent_id": GetCurrentUserID(c)}), &user)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found {
		curd.Error(c, 404, "数据不存在")
		return
	}

	// 仅变更所属组织时，直接同步成员关系
	if len(payload) == 0 {
		if !orgProvided {
			curd.Error(c, 400, "没有可更新的字段")
			return
		}
		if err := syncUserOrg(user.ID, orgID); err != nil {
			curd.Error(c, 500, err.Error())
			return
		}
		curd.Success(c, nil)
		return
	}

	// 交给通用更新逻辑
	body, _ := json.Marshal(payload)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	h.Handler.Update(c)
	if orgProvided {
		_ = syncUserOrg(user.ID, orgID)
	}
}

// Search 账号列表：附带所属组织（组织关系来自 org_member）
func (h *AccountHandler) Search(c *gin.Context) {
	var params curd.ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		curd.Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	conds := []builder.Cond{}
	if pc := params.BuildCond(); pc != nil {
		conds = append(conds, pc)
	}
	if params.HasKeyword() && len(h.Keywords) > 0 {
		kws := make([]builder.Cond, 0, len(h.Keywords))
		for _, kw := range h.Keywords {
			kws = append(kws, builder.Like{kw, "%" + params.Keyword + "%"})
		}
		conds = append(conds, builder.Or(kws...))
	}
	conds = append(conds, builder.Eq{"parent_id": GetCurrentUserID(c)})

	var cond builder.Cond
	if len(conds) > 1 {
		cond = builder.And(conds...)
	} else {
		cond = conds[0]
	}

	total, err := db.Count(h.Collection, cond)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "id"
	}
	users := make([]types.User, 0, params.PageSize)
	if _, err := db.PagedFind(h.Collection, cond, params.Page, params.PageSize, orderBy, params.Order != "desc", &users); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	orgIDs := userOrgIDs(ids)
	items := make([]AccountUser, 0, len(users))
	for _, u := range users {
		items = append(items, AccountUser{User: u, OrgIDs: orgIDs[u.ID]})
	}
	curd.SuccessTotal(c, items, total)
}

// Delete 删除子账号：同时清理组织成员关系
func (h *AccountHandler) Delete(c *gin.Context) {
	var user types.User
	found, err := db.FindOne("user", builder.And(builder.Eq{"id": c.Param("id")}, builder.Eq{"parent_id": GetCurrentUserID(c)}), &user)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found {
		curd.Error(c, 404, "数据不存在")
		return
	}
	_, _ = db.DeleteOne("org_member", builder.Eq{"user_id": user.ID})
	h.Handler.Delete(c)
}

// Get 账号详情：附带所属组织（组织关系来自 org_member）
func (h *AccountHandler) Get(c *gin.Context) {
	var user types.User
	found, err := db.FindOne(h.Collection, builder.And(builder.Eq{"id": c.Param("id")}, builder.Eq{"parent_id": GetCurrentUserID(c)}), &user)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if !found {
		curd.Error(c, 404, "数据不存在")
		return
	}
	curd.Success(c, AccountUser{User: user, OrgIDs: userOrgIDs([]string{user.ID})[user.ID]})
}
