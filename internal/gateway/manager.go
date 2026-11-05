package gateway

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/protocol"
	_ "github.com/god-jason/iot-master/pkg/protocol/cjt188"
	_ "github.com/god-jason/iot-master/pkg/protocol/dlt645"
	_ "github.com/god-jason/iot-master/pkg/protocol/iec"
	_ "github.com/god-jason/iot-master/pkg/protocol/jt808"
	_ "github.com/god-jason/iot-master/pkg/protocol/modbus"

	"xorm.io/builder"
)

// startCheckTimeout 启动后等待监听结果的最长时间
//
// 端口占用等启动失败会在此窗口内由 Listen 返回错误；超时则视为监听已正常开始。
const startCheckTimeout = 200 * time.Millisecond

// GatewayServerInstance 一个运行中的 Gateway Server 实例
type GatewayServerInstance struct {
	Config  *types.GatewayServer // DB 中的网关配置（启停与设备加载依据）
	Server  *Server              // 实际监听与连接处理的 TCP 服务器
	running atomic.Bool          // 是否处于运行态（Listen 返回后置 false）
}

// IsRunning 服务器是否在运行
func (i *GatewayServerInstance) IsRunning() bool { return i.running.Load() }

// GatewayManager 管理多个 Gateway Server 实例
type GatewayManager struct {
	mu        sync.RWMutex                     // 保护 instances
	instances map[int64]*GatewayServerInstance // serverID → instance
}

var defaultManager = &GatewayManager{instances: make(map[int64]*GatewayServerInstance)}

// GetGatewayManager 获取全局 GatewayManager 单例
func GetGatewayManager() *GatewayManager { return defaultManager }

// GetInstance 按 serverID 获取运行中的实例
func (m *GatewayManager) GetInstance(serverID int64) *GatewayServerInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.instances[serverID]
}

// ListInstances 列出所有运行中的实例
func (m *GatewayManager) ListInstances() []*GatewayServerInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*GatewayServerInstance, 0, len(m.instances))
	for _, inst := range m.instances {
		list = append(list, inst)
	}
	return list
}

// StartServer 根据 DB 配置启动一个 Gateway Server 实例
func (m *GatewayManager) StartServer(cfg *types.GatewayServer) (*GatewayServerInstance, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gateway: config is nil")
	}

	// 已运行实例直接复用，避免同一 serverID 重复监听
	// 检查与登记在同一把锁内完成，并发调用只会有一个实例真正监听同一端口
	m.mu.Lock()
	if existing, ok := m.instances[cfg.ID]; ok && existing.IsRunning() {
		m.mu.Unlock()
		return existing, nil
	}

	handler := &gatewayHandler{}

	// 配置里的协议名未注册时退回默认协议（协议名在包内已验证可用）
	opts := []ServerOption{
		WithServerID(cfg.ID),
		WithHandler(handler),
		WithConcurrency(cfg.Concurrency),
		WithInterval(parseInterval(cfg.Interval)),
	}
	if protocolAvailable(cfg.Protocol) {
		opts = append(opts, WithDefaultProtocol(cfg.Protocol))
	}
	srv := NewServer(opts...)

	// 先登记实例再启动监听，保证注册回调能通过 serverID 找到 Server
	inst := &GatewayServerInstance{
		Config: cfg,
		Server: srv,
	}
	inst.running.Store(true)
	handler.inst = inst

	m.instances[cfg.ID] = inst
	m.mu.Unlock()

	// 统一监听全部网卡（0.0.0.0），忽略历史配置中的 host，保证端口对外全监听
	addr := "0.0.0.0:" + strconv.Itoa(cfg.Port)

	// errCh 缓冲 1，Listen 结束后的结果一定可写入，不会阻塞清理 goroutine
	errCh := make(chan error, 1)

	// Listen 阻塞直到引擎退出
	go func() {
		logger.Info("软网关 Server 启动中", "server_id", cfg.ID, "name", cfg.Name, "addr", addr)
		err := srv.Listen("tcp://" + addr)
		inst.running.Store(false)

		// 仅当映射仍指向本实例时才移除，避免误删重启后的新实例
		m.mu.Lock()
		if cur, ok := m.instances[cfg.ID]; ok && cur == inst {
			delete(m.instances, cfg.ID)
		}
		m.mu.Unlock()

		// Listen 非 nil 错误表示未进入运行态（如端口被占用）
		if err != nil {
			logger.Error("软网关 Server 运行结束", "server_id", cfg.ID, "error", err)
		}
		errCh <- err
	}()

	// 短暂等待监听结果：启动失败时实例已清理，错误返回给调用方（API 可提示用户）
	// 超时视为监听已正常开始，不阻塞调用方
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("gateway: 监听 %s 失败: %w", addr, err)
		}
	case <-time.After(startCheckTimeout):
	}

	return inst, nil
}

// StopServer 停止指定 Server 实例
func (m *GatewayManager) StopServer(serverID int64) {
	// 先从管理器摘除，后续启动请求不会复用正在关闭的实例
	m.mu.Lock()
	inst, ok := m.instances[serverID]
	if ok {
		inst.running.Store(false)
		delete(m.instances, serverID)
	}
	m.mu.Unlock()

	// Stop 在锁外执行：引擎关闭可能阻塞，避免持锁影响其它实例操作
	if ok && inst.Server != nil {
		logger.Info("软网关 Server 关闭", "server_id", serverID, "name", inst.Config.Name)
		inst.Server.Stop()
	}
}

// ShutdownAll 关闭所有 Server 实例（boot 关闭时调用）
func (m *GatewayManager) ShutdownAll() {
	for _, inst := range m.ListInstances() {
		m.StopServer(inst.Config.ID)
	}
}

// ReloadDevices 按 link_id 重新加载 DB 设备并挂载到运行中的链路。
//
// 用于设备配置（新增/修改/删除）变更后的热更新：
// 先注销该链路原有的 iot 设备，再按最新 DB 记录 Mount 并重新注册。
// 链路不在运行态时不做处理（下次注册时由 OnRegister 按 DB 加载）。
func (i *GatewayServerInstance) ReloadDevices(linkID string) []*Device {
	if i == nil || i.Server == nil || linkID == "" {
		return nil
	}

	var link *Link
	for _, l := range i.Server.Links() {
		if l.ID() == linkID {
			link = l
			break
		}
	}
	if link == nil || !link.IsRegistered() {
		return nil
	}

	for _, old := range link.Devices() {
		if old != nil {
			iot.Unregister(old.ID())
		}
	}

	devs := i.Server.MountDevices(link, loadLinkDevices(linkID))
	for _, dev := range devs {
		dev.SetOnline()
		iot.Register(dev.ID(), dev)
	}
	logger.Info("软网关链路设备已热更新", "link_id", linkID, "count", len(devs))
	return devs
}

// ============ DB 查询 ============

// createDefaultDevices 按 GatewayServer.DefaultDevices 为新注册的链路创建设备记录
//
// 每个默认设备生成一条 device 行并绑定 link_id；协议地址取自 GatewayDevice.Address
// （为空时按序号生成），SN 由 link_id + 地址组成以保证唯一。
func createDefaultDevices(cfg *types.GatewayServer, link *Link) []*types.Device {
	if cfg == nil || len(cfg.DefaultDevices) == 0 {
		return nil
	}

	now := time.Now()
	out := make([]*types.Device, 0, len(cfg.DefaultDevices))
	for i, d := range cfg.DefaultDevices {
		if d.ModelID == "" {
			continue
		}
		addr := d.Address
		if addr == "" {
			addr = strconv.Itoa(i + 1)
		}

		row := &types.Device{
			ID:      db.NewID(),
			SN:      link.ID() + "-" + addr,
			Name:    d.Name,
			ModelID: d.ModelID,
			Address: addr,
			LinkID:  link.ID(),
			Created: now,
		}
		if _, err := db.InsertOne("device", row); err != nil {
			logger.Warn("软网关创建默认设备失败", "link_id", link.ID(), "name", d.Name, "error", err)
			continue
		}
		logger.Info("软网关已创建默认设备", "link_id", link.ID(), "device_id", row.ID, "name", d.Name, "address", addr, "model", d.ModelID)
		out = append(out, row)
	}
	return out
}

// ensureLinkRecord 确保链路在 link 表中有记录（ID 即注册包 SN）
func ensureLinkRecord(serverID int64, link *Link) {
	id := link.ID()
	if id == "" {
		return
	}

	var row types.Link
	found, err := db.FindOne("link", builder.Eq{"id": id}, &row)
	if err != nil {
		logger.Warn("软网关查询 link 失败", "link_id", id, "error", err)
		return
	}
	if found {
		return
	}

	row = types.Link{ID: id, ServerID: serverID, Name: link.Name()}
	if _, err := db.InsertOne("link", &row); err != nil {
		logger.Warn("软网关创建 link 失败", "link_id", id, "error", err)
	}
}

// loadLinkDevices 按 Device.LinkID 查询该链路下要挂载的设备
func loadLinkDevices(linkID string) []*types.Device {
	if linkID == "" {
		return nil
	}
	// 只加载未禁用设备；上限 10000 条，避免单链路设备过多拖慢注册
	var rows []types.Device
	err := db.Find("device", builder.Eq{"link_id": linkID, "disabled": false}, &rows, 0, 10000)
	if err != nil {
		logger.Warn("软网关加载链路设备失败", "link_id", linkID, "error", err)
		return nil
	}
	out := make([]*types.Device, 0, len(rows))
	for i := range rows {
		out = append(out, &rows[i])
	}
	return out
}

// parseInterval 解析 "5s" 等间隔字符串，非法/为空返回 0
func parseInterval(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0
	}
	return d
}

// ============ handler 实现 ============

// gatewayHandler 软网关业务回调：按链路加载设备、注册/注销 iot
type gatewayHandler struct {
	inst *GatewayServerInstance // 所属实例，用于访问 Server 与配置
}

// OnRegister 链路注册：链路落库、按 Device.LinkID 加载设备并逐一 Mount、注册 iot
func (h *gatewayHandler) OnRegister(ctx context.Context, link *Link) {
	if link == nil {
		return
	}
	logger.Info("软网关链路注册", "link_id", link.ID(), "conn_id", link.ConnID, "sn", link.SN())
	if h.inst == nil || h.inst.Server == nil {
		return
	}

	// 1. 链路落库（设备通过 Device.LinkID 关联）
	ensureLinkRecord(h.inst.Config.ID, link)

	// 2. 禁用的链路直接拒绝连接
	var row types.Link
	found, err := db.FindOne("link", builder.Eq{"id": link.ID()}, &row)
	if err != nil {
		logger.Warn("软网关查询链路配置失败", "link_id", link.ID(), "error", err)
	} else if found && row.Disabled {
		logger.Warn("软网关链路已禁用，拒绝连接", "link_id", link.ID(), "conn_id", link.ConnID)
		if link.Conn != nil {
			_ = link.Conn.Close()
		}
		return
	}

	// 3. 按 Device.LinkID 查询设备；新链路没有设备时按默认设备配置自动创建
	rows := loadLinkDevices(link.ID())
	if len(rows) == 0 {
		rows = createDefaultDevices(h.inst.Config, link)
	}
	// 4. 逐一 Mount 到协议实例
	devs := h.inst.Server.MountDevices(link, rows)

	// 5. 注册到 iot
	for _, dev := range devs {
		dev.SetOnline()
		iot.Register(dev.ID(), dev)
		logger.Info("软网关设备已注册到 iot", "device_id", dev.ID(), "sn", dev.SN, "link_id", link.ID())
	}
}

// OnDisconnect 链路断开：注销该链路下的所有 iot 设备
func (h *gatewayHandler) OnDisconnect(ctx context.Context, link *Link, reason string) {
	if link == nil {
		return
	}
	logger.Info("软网关链路断开", "link_id", link.ID(), "conn_id", link.ConnID, "reason", reason)

	for _, dev := range link.Devices() {
		if dev == nil {
			continue
		}
		iot.Unregister(dev.ID())
		logger.Info("软网关设备已从 iot 移除", "device_id", dev.ID())
	}
}

// protocolAvailable 判断协议名是否已注册
func protocolAvailable(name string) bool {
	if name == "" {
		return false
	}
	// 空名或清单中无工厂方法（未注册/初始化失败）均视为不可用
	ma := protocol.GetManifest(name)
	return ma.Factory != nil
}
