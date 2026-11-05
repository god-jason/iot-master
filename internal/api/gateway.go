package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/gateway"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/curd"
	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// ============ GatewayServer CRUD ============

// gatewayListServers 查询全部网关服务器配置
// 列表项额外附带运行状态、当前连接数与轮询调度项数，方便前端一屏展示
func gatewayListServers(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var servers []types.GatewayServer
	// 分页参数可选：未传 page_size 时保持原有行为返回全部配置
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0"))
	if page <= 0 {
		page = 1
	}
	if pageSize < 0 || pageSize > 100 {
		pageSize = 0
	}
	var total int64
	var err error
	if pageSize > 0 {
		total, err = db.PagedFind("gateway_server", builder.NewCond(), page, pageSize, "id", false, &servers)
	} else {
		err = db.Engine.Find(&servers)
		total = int64(len(servers))
	}
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 汇总运行实例的启停状态（只有已启动的服务器才会存在实例）
	runningMap := make(map[int64]bool)
	for _, inst := range gateway.GetGatewayManager().ListInstances() {
		runningMap[inst.Config.ID] = inst.IsRunning()
	}

	// 逐条拼接数据库配置与运行时指标
	items := make([]gin.H, 0, len(servers))
	for _, s := range servers {
		items = append(items, gin.H{
			"id":              s.ID,
			"name":            s.Name,
			"host":            s.Host,
			"port":            s.Port,
			"protocol":        s.Protocol,
			"interval":        s.Interval,
			"concurrency":     s.Concurrency,
			"disabled":        s.Disabled,
			"running":         runningMap[s.ID],
			"conn_count":      serverConnCount(s.ID),
			"poll_entries":    serverPollEntries(s.ID),
			"default_devices": s.DefaultDevices,
		})
	}
	curd.Success(c, gin.H{"items": items, "total": total})
}

// gatewayGetServer 按 ID 查询单台网关服务器配置
func gatewayGetServer(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的服务器ID")
		return
	}
	var s types.GatewayServer
	ok, err := db.FindOne("gateway_server", builder.Eq{"id": id}, &s)
	if err != nil || !ok {
		curd.Error(c, 404, "服务器不存在")
		return
	}
	curd.Success(c, s)
}

// gatewayCreateServer 新建网关服务器配置（port 必填，concurrency 缺省为 4）
func gatewayCreateServer(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var body types.GatewayServer
	if err := c.ShouldBindJSON(&body); err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	if body.Port == 0 {
		curd.Error(c, 400, "port 必填")
		return
	}
	// 未指定并发数时使用默认值 4
	if body.Concurrency == 0 {
		body.Concurrency = 4
	}
	body.Created = time.Now()
	body.Updated = time.Now()
	id, err := db.InsertOne("gateway_server", &body)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	body.ID = id
	curd.Success(c, body)
}

// gatewayUpdateServer 按 ID 更新网关服务器配置（整字段覆盖，不重载运行中的实例）
func gatewayUpdateServer(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的服务器ID")
		return
	}
	var body types.GatewayServer
	if err := c.ShouldBindJSON(&body); err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	body.ID = id
	body.Updated = time.Now()
	_, err = db.UpdateOne("gateway_server", builder.Eq{"id": id}, map[string]interface{}{
		"name":            body.Name,
		"host":            body.Host,
		"port":            body.Port,
		"protocol":        body.Protocol,
		"interval":        body.Interval,
		"concurrency":     body.Concurrency,
		"disabled":        body.Disabled,
		"default_devices": body.DefaultDevices,
		"updated":         body.Updated,
	})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// gatewayDeleteServer 删除网关服务器配置（先停止运行实例，再删除数据库记录）
func gatewayDeleteServer(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的服务器ID")
		return
	}
	// 先停止实例，避免删除配置后后台仍持有监听端口与连接
	gateway.GetGatewayManager().StopServer(id)
	_, err = db.DeleteOne("gateway_server", builder.Eq{"id": id})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, nil)
}

// gatewayStartServer 启动指定网关服务器（按数据库配置创建运行实例）
func gatewayStartServer(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的服务器ID")
		return
	}
	var s types.GatewayServer
	ok, err := db.FindOne("gateway_server", builder.Eq{"id": id}, &s)
	if err != nil || !ok {
		curd.Error(c, 404, "服务器不存在")
		return
	}
	// 先确认配置存在，再交由管理器创建并启动实例
	inst, err := gateway.GetGatewayManager().StartServer(&s)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	curd.Success(c, gin.H{"server_id": inst.Config.ID})
}

// gatewayStopServer 停止指定网关服务器的运行实例（不删除数据库配置）
func gatewayStopServer(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的服务器ID")
		return
	}
	gateway.GetGatewayManager().StopServer(id)
	curd.Success(c, nil)
}

// ============ 辅助 ============

// serverConnCount 返回指定服务器当前的活跃链路数（未运行时为 0）
func serverConnCount(serverID int64) int {
	inst := gateway.GetGatewayManager().GetInstance(serverID)
	if inst == nil || inst.Server == nil {
		return 0
	}
	return len(inst.Server.Links())
}

// serverPollEntries 返回指定服务器的轮询调度项数（底层按当前链路数统计，未运行时为 0）
func serverPollEntries(serverID int64) int {
	inst := gateway.GetGatewayManager().GetInstance(serverID)
	if inst == nil || inst.Server == nil {
		return 0
	}
	return inst.Server.SchedulerEntries()
}

// ============ 连接管理 ============

// getServerInstance 解析路径中的服务器 ID 并取出其运行实例
// 解析失败或实例未运行时已直接写出错误响应，调用方只需在返回 false 时返回
func getServerInstance(c *gin.Context) (*gateway.GatewayServerInstance, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的服务器ID")
		return nil, false
	}
	inst := gateway.GetGatewayManager().GetInstance(id)
	if inst == nil {
		curd.Error(c, 404, "服务器未运行")
		return nil, false
	}
	return inst, true
}

// gatewayListConnections 查询单台服务器下的全部链路连接（含链路下的设备）
func gatewayListConnections(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	inst, ok := getServerInstance(c)
	if !ok {
		return
	}
	items := collectConnectionItems(inst)
	curd.Success(c, gin.H{"items": items})
}

// gatewayListAllConnections 汇总当前所有运行实例的链路连接
func gatewayListAllConnections(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var items []gin.H
	for _, inst := range gateway.GetGatewayManager().ListInstances() {
		items = append(items, collectConnectionItems(inst)...)
	}
	curd.Success(c, gin.H{"items": items})
}

// collectConnectionItems 把实例下的运行链路转换为响应项集合（实例为空时返回 nil）
func collectConnectionItems(inst *gateway.GatewayServerInstance) []gin.H {
	if inst == nil || inst.Server == nil {
		return nil
	}
	links := inst.Server.Links()
	items := make([]gin.H, 0, len(links))
	for _, l := range links {
		items = append(items, linkItem(l))
	}
	return items
}

// linkItem 一条链路的运行状态（含其下设备）
func linkItem(l *gateway.Link) gin.H {
	devs := l.Devices()
	devItems := make([]gin.H, 0, len(devs))
	for _, d := range devs {
		devItems = append(devItems, gin.H{
			"id":        d.ID(),
			"sn":        d.SN,
			"name":      d.Name,
			"model_id":  d.ModelID(),
			"mounted":   d.Mounted(),
			"online":    d.OnlineStatus(),
			"last_seen": d.LastSeenAt().Format(time.RFC3339),
		})
	}

	pollInfo := gin.H{}
	if pc := l.PollConfig(); pc != nil {
		pollInfo = gin.H{"interval": pc.Interval.String()}
	}

	registeredAt := ""
	if t := l.RegisteredAt(); !t.IsZero() {
		registeredAt = t.Format(time.RFC3339)
	}
	lastSeen := ""
	if t := l.LastSeen(); !t.IsZero() {
		lastSeen = t.Format(time.RFC3339)
	}

	return gin.H{
		"conn_id":       l.ConnID,
		"server_id":     l.ServerID,
		"link_id":       l.ID(),
		"sn":            l.SN(),
		"imei":          l.IMEI(),
		"remote":        l.Remote,
		"protocol":      l.Protocol,
		"registered":    l.IsRegistered(),
		"registered_at": registeredAt,
		"last_seen":     lastSeen,
		"poll":          pollInfo,
		"devices":       devItems,
	}
}

// gatewayConnectionDetail 查询单条链路连接的详情（含链路下的设备）
func gatewayConnectionDetail(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	inst, ok := getServerInstance(c)
	if !ok {
		return
	}
	// conn_id 是运行时连接的标识，与链路 ID（注册包 SN）不是同一概念
	connID, err := strconv.ParseInt(c.Param("conn_id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的连接ID")
		return
	}
	l, ok := inst.Server.GetLink(connID)
	if !ok {
		curd.Error(c, 404, "连接不存在")
		return
	}
	curd.Success(c, linkItem(l))
}

// gatewaySetPoll 设置或清除指定链路的轮询采集周期
func gatewaySetPoll(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	inst, ok := getServerInstance(c)
	if !ok {
		return
	}
	connID, err := strconv.ParseInt(c.Param("conn_id"), 10, 64)
	if err != nil {
		curd.Error(c, 400, "无效的连接ID")
		return
	}
	var body struct {
		Interval string `json:"interval"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		curd.Error(c, 400, err.Error())
		return
	}

	// interval 为空表示关闭该链路的轮询
	if body.Interval == "" {
		inst.Server.SetPoll(connID, nil)
	} else {
		// 解析 Go duration 字符串（如 5s、1m）后下发给链路
		dur, perr := time.ParseDuration(body.Interval)
		if perr != nil {
			curd.Error(c, 400, "interval 格式错误（如 5s）")
			return
		}
		inst.Server.SetPoll(connID, &gateway.PollConfig{Interval: dur})
	}
	curd.Success(c, nil)
}

// ============ 链路设备管理（Device.LinkID） ============
//
// 软网关的设备归属由 Device.LinkID 决定（Link ID = 注册包 SN），
// 与运行中的 TCP 连接（conn_id）无关，因此这里以 link_id 作为路径参数。
// 设备变更后若链路正在运行，会触发 ReloadDevices 热更新挂载。

// gatewayListLinkDevices 分页查询某链路下挂载的设备
// 结果同时合并物模型信息与链路运行态（挂载/在线/最后上报时间）
func gatewayListLinkDevices(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	linkID := c.Param("link_id")
	if linkID == "" {
		curd.Error(c, 400, "无效的链路ID")
		return
	}

	// 分页参数兜底：page >= 1，page_size 限制在 1~100
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var cond builder.Cond = builder.Eq{"link_id": linkID}
	// 关键词同时匹配设备名称与 SN
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		cond = builder.And(cond, builder.Or(
			builder.Like{"name", "%" + kw + "%"},
			builder.Like{"sn", "%" + kw + "%"},
		))
	}

	var rows []types.Device
	total, err := db.PagedFind("device", cond, page, pageSize, "id", false, &rows)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	// 链路正在运行时，收集各设备的实时状态供列表补充
	link, _ := findRuntimeLink(linkID)
	state := make(map[string]gin.H)
	if link != nil {
		for _, d := range link.Devices() {
			state[d.ID()] = gin.H{
				"mounted":   d.Mounted(),
				"online":    d.OnlineStatus(),
				"last_seen": d.LastSeenAt().Format(time.RFC3339),
			}
		}
	}

	// 批量拼接物模型信息
	modelMap := map[string]*types.Model{}
	for _, d := range rows {
		if d.ModelID == "" {
			continue
		}
		if _, ok := modelMap[d.ModelID]; ok {
			continue
		}
		var m types.Model
		if found, _ := db.FindOne("model", builder.Eq{"id": d.ModelID}, &m); found {
			modelMap[d.ModelID] = &m
		}
	}

	items := make([]gin.H, 0, len(rows))
	for _, d := range rows {
		item := gin.H{
			"id":       d.ID,
			"sn":       d.SN,
			"name":     d.Name,
			"model_id": d.ModelID,
			"address":  d.Address,
			"disabled": d.Disabled,
			"created":  d.Created,
		}
		if m, ok := modelMap[d.ModelID]; ok {
			item["model_name"] = m.Name
			item["model_icon"] = m.Icon
			item["model_type"] = m.Type
		}
		if st, ok := state[d.ID]; ok {
			item["mounted"] = st["mounted"]
			item["online"] = st["online"]
			item["last_seen"] = st["last_seen"]
		}
		items = append(items, item)
	}

	curd.Success(c, gin.H{"items": items, "total": total})
}

// gatewayCreateLinkDevice 在指定链路下创建设备（自动补 ID、SN 与 link_id）
func gatewayCreateLinkDevice(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	linkID := c.Param("link_id")
	if linkID == "" {
		curd.Error(c, 400, "无效的链路ID")
		return
	}

	// 校验链路存在，避免产生无归属的孤儿设备
	var row types.Link
	found, err := db.FindOne("link", builder.Eq{"id": linkID}, &row)
	if err != nil || !found {
		curd.Error(c, 404, "链路不存在")
		return
	}

	var body types.Device
	if err := c.ShouldBindJSON(&body); err != nil {
		curd.Error(c, 400, err.Error())
		return
	}
	if body.ModelID == "" {
		curd.Error(c, 400, "model_id 必填")
		return
	}
	if body.Address == "" {
		curd.Error(c, 400, "address 必填")
		return
	}

	// 未传 ID / SN 时按链路与设备地址生成默认值
	if body.ID == "" {
		body.ID = db.NewID()
	}
	if body.SN == "" {
		body.SN = fmt.Sprintf("%s-%s", linkID, body.Address)
	}
	body.LinkID = linkID
	body.Created = time.Now()

	if _, err := db.InsertOne("device", &body); err != nil {
		curd.Error(c, 500, err.Error())
		return
	}

	reloadLinkDevices(linkID)
	curd.Success(c, body)
}

// gatewayUpdateLinkDevice 更新链路下的设备（仅允许白名单字段）
func gatewayUpdateLinkDevice(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	linkID := c.Param("link_id")
	deviceID := c.Param("device_id")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		curd.Error(c, 400, err.Error())
		return
	}

	// 只提取允许更新的字段，其余字段静默忽略
	updateSet := map[string]interface{}{}
	for _, field := range []string{"sn", "name", "model_id", "address", "disabled"} {
		if v, ok := payload[field]; ok {
			updateSet[field] = v
		}
	}
	if len(updateSet) == 0 {
		curd.Error(c, 400, "没有可更新的字段")
		return
	}

	affected, err := db.UpdateOne("device", builder.Eq{"id": deviceID, "link_id": linkID}, updateSet)
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	// 影响行数为 0 说明该设备不属于此链路
	if affected == 0 {
		curd.Error(c, 404, "设备不存在")
		return
	}

	reloadLinkDevices(linkID)
	curd.Success(c, nil)
}

// gatewayDeleteLinkDevice 删除链路下的设备（限定 link_id，避免误删其他链路的设备）
func gatewayDeleteLinkDevice(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	linkID := c.Param("link_id")
	deviceID := c.Param("device_id")

	affected, err := db.DeleteOne("device", builder.Eq{"id": deviceID, "link_id": linkID})
	if err != nil {
		curd.Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		curd.Error(c, 404, "设备不存在")
		return
	}

	reloadLinkDevices(linkID)
	curd.Success(c, nil)
}

// findRuntimeLink 在所有运行实例中查找指定 ID 的运行态链路
func findRuntimeLink(linkID string) (*gateway.Link, *gateway.GatewayServerInstance) {
	for _, inst := range gateway.GetGatewayManager().ListInstances() {
		if inst.Server == nil {
			continue
		}
		for _, l := range inst.Server.Links() {
			if l.ID() == linkID {
				return l, inst
			}
		}
	}
	return nil, nil
}

// reloadLinkDevices 触发所有运行实例按最新 DB 记录重挂载该链路设备
func reloadLinkDevices(linkID string) {
	for _, inst := range gateway.GetGatewayManager().ListInstances() {
		inst.ReloadDevices(linkID)
	}
}

// registerGatewayRoutes 注册 Gateway API 路由
func registerGatewayRoutes(r *gin.RouterGroup) {
	g := r.Group("/gateway")
	{
		// Gateway Server DB CRUD + 启停  —— :id = server_id
		g.GET("/servers", gatewayListServers)
		g.POST("/servers", gatewayCreateServer)
		g.GET("/servers/:id", gatewayGetServer)
		g.PUT("/servers/:id", gatewayUpdateServer)
		g.DELETE("/servers/:id", gatewayDeleteServer)
		g.POST("/servers/:id/start", gatewayStartServer)
		g.POST("/servers/:id/stop", gatewayStopServer)

		// 运行时链路管理（设备通过 Device.LinkID 归属链路）
		g.GET("/servers/:id/connections", gatewayListConnections)
		g.GET("/connections", gatewayListAllConnections)
		g.GET("/servers/:id/connections/:conn_id", gatewayConnectionDetail)
		g.PUT("/servers/:id/connections/:conn_id/poll", gatewaySetPoll)

		// 链路设备管理（link_id = 注册包 SN，设备通过 Device.LinkID 归属）
		g.GET("/links/:link_id/devices", gatewayListLinkDevices)
		g.POST("/links/:link_id/devices", gatewayCreateLinkDevice)
		g.PUT("/links/:link_id/devices/:device_id", gatewayUpdateLinkDevice)
		g.DELETE("/links/:link_id/devices/:device_id", gatewayDeleteLinkDevice)
	}
}
