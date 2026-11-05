// Package gateway 软网关：基于 gnet 的 TCP 服务器，负责链路注册、设备挂载与协议轮询采集。
//
// 职责边界：
//   - Server 负责监听、连接生命周期与注册包解析，每条 TCP 连接对应一条 Link；
//   - Link 持有该连接独立的 protocol.Protocol 实例与拼包缓冲，注册成功后按 Device.LinkID 加载并挂载设备；
//   - Poller 周期调度已挂载设备的协议轮询，采集结果写入设备属性（Device.ReportProperties）；
//   - GatewayManager 从 DB 配置启停多个 Server 实例。
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/protocol"
)

const (
	// maxTextFrame 纯文本注册包/心跳包的最大长度（超过则视为已完整的一行）
	maxTextFrame = 256
	// maxRegisterFrame JSON 注册包的最大长度
	maxRegisterFrame = 4096
	// pollTimeout 单次协议轮询的整体超时预算
	pollTimeout = 5 * time.Second
)

var (
	// errInvalidRegister 注册包内容非法
	errInvalidRegister = errors.New("gateway: 注册包非法")
	// errUnregisteredFrame 未注册连接直接发送协议帧
	errUnregisteredFrame = errors.New("gateway: 收到协议帧但连接未注册")
)

// ServerOption Server 构造选项
type ServerOption func(*Server)

// Server 基于 gnet 的软网关 TCP 服务器
//
// 每个 TCP 连接在 OnOpen 时创建一个独立的 protocol.Protocol 实例，
// 注册包解析成功后该连接成为 Link；Server 只保存 links，设备由 Link 内部持有。
type Server struct {
	gnet.BuiltinEventEngine

	engine  gnet.Engine     // gnet 引擎句柄，OnBoot 时注入，Stop 时用于关闭
	handler ProtocolHandler // 业务回调实现，默认 defaultHandler（空实现）
	poller  *Poller         // 轮询器，OnBoot 时创建并随 Server 生命周期运行

	mu    sync.RWMutex    // 保护 links
	links map[int64]*Link // connID → 连接上下文（仅存活连接）

	counter      uint64        // 连接自增序号，用于生成 connID
	serverID     int64         // Server 实例 ID（对应 gateway_server 表主键）
	defaultProto string        // 默认协议名，注册包未指定协议时使用
	interval     time.Duration // 默认轮询间隔，子设备未单独配置时生效
	concurrency  int           // 轮询 worker 数量
	tickInterval time.Duration // Poller 的 tick 周期
}

// NewServer 创建软网关服务器
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		links:        make(map[int64]*Link),
		handler:      defaultHandler{},
		defaultProto: protocol.TCP,
		concurrency:  4,
		tickInterval: time.Second,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithHandler 注入业务回调
func WithHandler(h ProtocolHandler) ServerOption {
	return func(s *Server) {
		if h != nil {
			s.handler = h
		}
	}
}

// WithDefaultProtocol 设置默认协议名（注册包未指定时使用）
func WithDefaultProtocol(p string) ServerOption {
	return func(s *Server) {
		if p != "" {
			s.defaultProto = p
		}
	}
}

// WithServerID 设置 Server 实例 ID
func WithServerID(id int64) ServerOption {
	return func(s *Server) { s.serverID = id }
}

// WithConcurrency 设置轮询 worker 数量（默认 4）
func WithConcurrency(n int) ServerOption {
	return func(s *Server) {
		if n > 0 {
			s.concurrency = n
		}
	}
}

// WithTickInterval 设置 Poller tick 周期（默认 1s）
func WithTickInterval(d time.Duration) ServerOption {
	return func(s *Server) {
		if d > 0 {
			s.tickInterval = d
		}
	}
}

// WithInterval 设置默认轮询间隔（子设备未单独配置时生效）
func WithInterval(d time.Duration) ServerOption {
	return func(s *Server) { s.interval = d }
}

// Listen 启动服务器（阻塞）
func (s *Server) Listen(addr string) error {
	return gnet.Run(s, addr, gnet.WithReusePort(true))
}

// Stop 主动停止
func (s *Server) Stop() {
	// Validate 校验引擎已初始化且未处于关闭流程，未启动时调用不会 panic
	if err := s.engine.Validate(); err == nil {
		_ = s.engine.Stop(context.Background())
	}
}

// ServerID 返回服务器实例 ID
func (s *Server) ServerID() int64 { return s.serverID }

// Interval 返回默认轮询间隔
func (s *Server) Interval() time.Duration { return s.interval }

// ============ gnet EventHandler ============

// OnBoot gnet 启动
func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	// 保存引擎句柄，供 Stop 使用
	s.engine = eng
	// 引擎启动后再创建轮询器：此时 links 已可安全遍历
	s.poller = NewPoller(s, s.concurrency, s.tickInterval, s.interval)
	go s.poller.Run()
	logger.Info("软网关 Server 启动", "server_id", s.serverID, "workers", s.concurrency, "tick", s.tickInterval)
	return gnet.None
}

// OnShutdown gnet 关闭
func (s *Server) OnShutdown(gnet.Engine) {
	if s.poller != nil {
		s.poller.Stop()
	}
	logger.Info("软网关 Server 已关闭", "server_id", s.serverID)
}

// OnOpen 新连接进来：创建 Link 并为其创建独立的协议实例
func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	// connID 全进程递增，作为 links 的 key 与日志关联标识
	id := atomic.AddUint64(&s.counter, 1)

	// 每条连接必须有独立协议实例（协议内部持有请求-响应状态与拼包缓冲）
	proto, ok := protocol.NewProtocol(s.defaultProto)
	if !ok {
		logger.Error("软网关无法创建协议实例", "server_id", s.serverID, "protocol", s.defaultProto)
		return nil, gnet.Close
	}

	link := newLink(s.serverID, int64(id), c, proto, s.defaultProto)
	// 把 Link 绑定到连接上下文，OnTraffic/OnClose 通过 getLink 取回
	c.SetContext(link)

	s.mu.Lock()
	s.links[int64(id)] = link
	s.mu.Unlock()

	// 协议主动上报（非轮询触发）的数据回调：不在 event loop 中处理，直接异步写属性
	onValues := func(device string, values map[string]any) {
		go s.reportAsync(link, device, values)
	}
	// 绑定连接的写通道与主动上报回调；失败时协议实例不可用，但连接仍保持由注册流程处理
	_ = proto.Attach(&gnetConnAdapter{c: c}, onValues)

	logger.Info("软网关连接建立", "server_id", s.serverID, "conn_id", id, "remote", c.RemoteAddr(), "protocol", s.defaultProto)
	return nil, gnet.None
}

// reportAsync 协议主动上报的数据写入对应设备属性
func (s *Server) reportAsync(link *Link, device string, values map[string]any) {
	if len(values) == 0 {
		return
	}
	// 按 deviceID（= DB 设备 ID）在链路设备中匹配，仅写入对应设备
	for _, dev := range link.Devices() {
		if dev != nil && dev.DeviceID() == device {
			_ = dev.ReportProperties(values, time.Now())
			return
		}
	}
}

// OnTraffic 收数据：注册包/心跳包解析 + 协议拼包解码
func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	link := getLink(c)
	n := c.InboundBuffered()
	if n <= 0 {
		// 没有可读数据，等待下一次事件
		return gnet.None
	}
	// Peek 只读不消费，配合 Discard 把 gnet 缓冲搬进 Link 自己的拼包缓冲
	buf, err := c.Peek(n)
	if err != nil {
		logger.Warn("软网关读取缓冲失败", "server_id", s.serverID, "conn_id", link.ConnID, "error", err)
		return gnet.None
	}
	// 数据已完整拷贝到 Link.buf，Discard 失败只影响 gnet 缓冲释放
	if _, err := c.Discard(n); err != nil {
		logger.Warn("软网关释放缓冲失败", "server_id", s.serverID, "conn_id", link.ConnID, "error", err)
	}

	link.buf = append(link.buf, buf...)

	// 循环消费缓冲：未注册阶段先解析注册包，注册后进入协议拼包
	for len(link.buf) > 0 {
		if !link.Registered {
			consumed, registered, rerr := s.consumeRegistration(link, link.buf)
			if rerr != nil {
				// 注册包非法（含未注册就发协议帧）：直接断开连接
				logger.Warn("软网关注册失败", "server_id", s.serverID, "conn_id", link.ConnID, "error", rerr)
				return gnet.Close
			}
			if consumed == 0 && !registered {
				return gnet.None // 等待更多注册数据
			}
			link.buf = link.buf[consumed:]
			continue
		}

		// 已注册：纯文本心跳包（仅当整行匹配心跳关键字才消费，避免误吞二进制帧）
		if line, consumed, ok := readLineFrame(link.buf); ok && isHeartbeat(string(line)) {
			link.buf = link.buf[consumed:]
			logger.Debug("软网关收到心跳包", "server_id", s.serverID, "conn_id", link.ConnID, "heartbeat", string(line))
			link.touch(time.Now())
			continue
		}

		// 交给协议实例拼包解码；ErrNeedMore（consumed<=0）表示当前数据不足一帧
		if link.Proto != nil {
			consumed, perr := link.Proto.OnData(link.buf)
			if perr != nil && !errors.Is(perr, protocol.ErrNeedMore) {
				logger.Debug("软网关协议解码错误", "server_id", s.serverID, "conn_id", link.ConnID, "error", perr)
			}
			if consumed <= 0 {
				break
			}
			// 防御协议实现返回越界消费值
			if consumed > len(link.buf) {
				consumed = len(link.buf)
			}
			link.buf = link.buf[consumed:]
		}

		// 每轮消费成功都刷新链路与设备的最近活跃时间
		link.touch(time.Now())
	}

	return gnet.None
}

// OnClose 连接断开
func (s *Server) OnClose(c gnet.Conn, err error) gnet.Action {
	link := getLink(c)

	// 先摘除协议实例再关闭，避免 Poller 等并发调用方取到已关闭实例
	if p := link.detachProtocol(); p != nil {
		_ = p.Close()
	}

	// 先摘除链接，避免 Poller 选中正在销毁的连接
	s.mu.Lock()
	delete(s.links, link.ConnID)
	s.mu.Unlock()

	// 仅已注册链路才通知业务层，未注册连接直接静默回收
	if link.IsRegistered() {
		link.setOffline()

		// 默认按对端关闭处理，有错误时以错误文本作为断开原因
		reason := "remote"
		if err != nil {
			reason = err.Error()
		}
		logger.Info("软网关连接断开", "server_id", s.serverID, "link_id", link.ID(), "conn_id", link.ConnID, "reason", reason)
		go s.handler.OnDisconnect(context.Background(), link, reason)
	}
	return gnet.None
}

// ============ 连接与设备访问 ============

// GetLink 按 connID 查连接上下文
func (s *Server) GetLink(connID int64) (*Link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.links[connID]
	return l, ok
}

// Links 所有连接上下文
func (s *Server) Links() []*Link {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Link, 0, len(s.links))
	for _, l := range s.links {
		list = append(list, l)
	}
	return list
}

// Devices 所有链路下的设备
func (s *Server) Devices() []*Device {
	var list []*Device
	for _, l := range s.Links() {
		list = append(list, l.Devices()...)
	}
	return list
}

// SchedulerEntries 当前轮询调度条目数（每条连接一个）
func (s *Server) SchedulerEntries() int { return len(s.Links()) }

// SetPoll 动态设置某连接的兜底轮询配置
func (s *Server) SetPoll(connID int64, poll *PollConfig) {
	l, ok := s.GetLink(connID)
	if !ok {
		return
	}
	l.SetPoll(poll)
}

// MountDevices 按 Link 加载到的 DB 设备逐一 Mount 到该连接的协议实例
//
// 设备点位（从站地址 / 功能码 / 起始地址 / 数量）由协议实例内部维护，
// 这里只传 deviceId、协议地址（Device.Address）与物模型。
func (s *Server) MountDevices(link *Link, rows []*types.Device) []*Device {
	if link == nil {
		return nil
	}
	// 协议实例可能已随连接关闭被摘除（本方法常在 OnRegister goroutine 中调用）
	proto := link.ProtocolInstance()
	if proto == nil {
		return nil
	}

	devs := make([]*Device, 0, len(rows))
	for _, row := range rows {
		// 跳过空行、禁用设备与未绑定物模型的设备
		if row == nil || row.Disabled || row.ModelID == "" {
			continue
		}
		// 物模型必须先完成协议归一化注册，否则协议实例无法解析点位
		if !ensureModelRegistered(row.ModelID) {
			logger.Warn("软网关设备挂载跳过：物模型未完成协议注册", "server_id", s.serverID, "link_id", link.ID(), "device_id", row.ID, "model_id", row.ModelID)
			continue
		}

		dev := newDevice(link, row)
		// 单个设备挂载失败不影响同链路其它设备
		if err := proto.Mount(dev.DeviceID(), row.Address, row.ModelID); err != nil {
			logger.Warn("软网关设备挂载失败", "server_id", s.serverID, "link_id", link.ID(), "device_id", row.ID, "error", err)
			continue
		}
		dev.setMounted(true)
		devs = append(devs, dev)
	}

	// 原子替换设备列表：新设备已挂载成功后再注销旧设备，避免中间态轮询
	old := link.replaceDevices(devs)
	for _, o := range old {
		if o != nil {
			proto.UnMount(o.DeviceID())
		}
	}
	// 设备集合变化后重置轮询时间，让下一轮调度尽快生效
	link.reschedule()

	logger.Info("软网关链路设备已挂载", "server_id", s.serverID, "link_id", link.ID(), "count", len(devs))
	return devs
}

// ============ 内部 ============

// gnetConnAdapter 让 gnet.Conn 满足 io.ReadWriteCloser
type gnetConnAdapter struct {
	c gnet.Conn // 底层 gnet 连接
}

// Write 实现 io.Writer：直接写回该 TCP 连接
func (a *gnetConnAdapter) Write(p []byte) (int, error) { return a.c.Write(p) }

// Read 实现 io.Reader：数据由 OnTraffic 读取并交给协议实例，这里恒返回 io.EOF
func (a *gnetConnAdapter) Read(p []byte) (int, error) { return 0, io.EOF }

// Close 实现 io.Closer：关闭底层连接
func (a *gnetConnAdapter) Close() error { return a.c.Close() }

// getLink 从 gnet 连接上下文中取出 Link
//
// OnOpen 必须为连接设置 *Link 上下文；取不到时返回 ConnID=-1 的空壳对象，
// 保证后续访问字段不 panic（该连接不会被注册）。
func getLink(c gnet.Conn) *Link {
	if v := c.Context(); v != nil {
		if l, ok := v.(*Link); ok {
			return l
		}
	}
	return &Link{ConnID: -1}
}

// consumeRegistration 尝试解析注册包，返回消费字节数与是否完成注册
func (s *Server) consumeRegistration(link *Link, buf []byte) (int, bool, error) {
	trimmed := bytes.TrimLeft(buf, " \t\r\n")
	if len(trimmed) == 0 {
		// 只有空白字符：视为已消费，继续等待有效注册内容
		return len(buf), false, nil
	}

	// JSON 注册包
	if trimmed[0] == '{' {
		pkt, consumed, ok, err := decodeJSONPacket(buf)
		if err != nil {
			return 0, false, err
		}
		if !ok {
			return 0, false, nil // 数据不完整，等待后续
		}
		// SN 缺失时按 device/IMEI 兜底，保证链路身份唯一
		if pkt.SN == "" {
			pkt.SN = pkt.Device
		}
		// 三种身份标识全为空则视为非法注册包
		if pkt.SN == "" && pkt.IMEI == "" {
			return 0, false, errInvalidRegister
		}
		s.registerLink(link, pkt)
		return consumed, true, nil
	}

	// 未注册连接直接发二进制协议帧 → 拒绝
	if !isPrintableByte(buf[0]) && looksLikeModbusFrame(buf) {
		return 0, false, errUnregisteredFrame
	}

	// 纯文本注册包（SN）
	line, consumed, ok := readTextFrame(buf)
	if !ok {
		// 缓冲已超过上限仍无法判定为一行，按非法注册包处理
		if len(buf) > maxRegisterFrame {
			return 0, false, errInvalidRegister
		}
		return 0, false, nil
	}

	sn := string(line)
	if sn == "" {
		return consumed, false, nil
	}
	// 注册前先收到心跳：消费掉但不改变注册状态
	if isHeartbeat(sn) {
		return consumed, false, nil
	}

	s.registerLink(link, RegisterPacket{SN: sn})
	return consumed, true, nil
}

// registerLink 注册包解析成功：确定链路身份并交给业务层加载设备
func (s *Server) registerLink(link *Link, pkt RegisterPacket) {
	link.register(pkt)

	logger.Info("软网关链路注册", "server_id", s.serverID, "link_id", link.ID(), "sn", link.SN(), "imei", link.IMEI(), "conn_id", link.ConnID, "protocol", link.Protocol, "remote", link.Remote)
	// 业务层回调需要访问 DB，放到 goroutine 中避免阻塞 gnet event loop
	go s.handler.OnRegister(context.Background(), link)
}

// decodeJSONPacket 解析 JSON 注册包，ok=false 表示数据不完整
func decodeJSONPacket(buf []byte) (RegisterPacket, int, bool, error) {
	var pkt RegisterPacket
	trimmed := bytes.TrimLeft(buf, " \t\r\n")

	dec := json.NewDecoder(bytes.NewReader(trimmed))
	if err := dec.Decode(&pkt); err != nil {
		// 数据被截断：未达上限则等待后续数据，达到上限则判定为非法注册包
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			if len(buf) >= maxRegisterFrame {
				return pkt, 0, false, errInvalidRegister
			}
			return pkt, 0, false, nil
		}
		return pkt, 0, false, fmt.Errorf("gateway: 注册包 JSON 解析失败: %w", err)
	}

	// 消费长度 = 前导空白 + 实际解析到的 JSON 字节数（其后可能紧跟协议帧）
	consumed := len(buf) - len(trimmed) + int(dec.InputOffset())
	return pkt, consumed, true, nil
}

// readTextFrame 从缓冲中截取一行纯文本（以 CR/LF 或首个不可打印字节为界）
// ok=false 表示当前缓冲还不足一行，需等待更多数据。
func readTextFrame(buf []byte) (line []byte, consumed int, ok bool) {
	for i := 0; i < len(buf); i++ {
		b := buf[i]
		// 优先以 CR/LF 作为行结束符
		if b == '\r' || b == '\n' {
			return bytes.TrimSpace(buf[:i]), i + 1, true
		}
		// 遇到不可打印字节：首字节即不可打印说明是二进制帧，否则以该位置为界截断
		if b < 0x20 || b > 0x7e {
			if i == 0 {
				return nil, 0, false // 二进制帧
			}
			return bytes.TrimSpace(buf[:i]), i, true
		}
	}
	// 无结束符：达到长度上限时按已收到内容作为一整行
	if len(buf) >= maxTextFrame {
		return bytes.TrimSpace(buf), len(buf), true
	}
	return nil, 0, false
}

// readLineFrame 已注册连接上的纯文本行（心跳包）：必须以 CR/LF 结尾
func readLineFrame(buf []byte) (line []byte, consumed int, ok bool) {
	// 必须以换行结尾，避免把二进制帧的前缀误判为文本心跳
	idx := bytes.IndexAny(buf, "\r\n")
	if idx < 0 {
		return nil, 0, false
	}
	return bytes.TrimSpace(buf[:idx]), idx + 1, true
}

// isPrintableByte 是否为可打印 ASCII 字节
func isPrintableByte(b byte) bool { return b >= 0x20 && b <= 0x7e }

// isHeartbeat 判断纯文本心跳包内容
func isHeartbeat(s string) bool {
	// 大小写不敏感，覆盖常见纯文本心跳关键字
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "heartbeat", "hb", "ping", "pong", "keepalive", "ka":
		return true
	}
	return false
}

// looksLikeModbusFrame 粗略判断是否 Modbus RTU 帧（用于拒绝未注册设备直接发协议帧）
func looksLikeModbusFrame(buf []byte) bool {
	if len(buf) < 2 {
		return false
	}
	// Modbus 从站地址范围 1~247，功能码 1~0x17
	addr, fc := buf[0], buf[1]
	return addr >= 1 && addr <= 247 && fc >= 1 && fc <= 0x17
}

// ensureModelRegistered 确保物模型已完成协议归一化注册
func ensureModelRegistered(modelID string) bool {
	// 已在 protocol 归一化缓存中，直接使用
	if _, ok := protocol.GetModel(modelID); ok {
		return true
	}
	m, ok := iot.GetModel(modelID)
	if !ok {
		logger.Warn("软网关物模型不存在", "model", modelID)
		return false
	}
	iot.SetModel(modelID, m) // 同时刷新 iot 缓存与 protocol 归一化缓存
	// 归一化后仍取不到，说明物模型结构不合法
	if _, ok := protocol.GetModel(modelID); !ok {
		logger.Warn("软网关物模型归一化失败", "model", modelID)
		return false
	}
	return true
}
