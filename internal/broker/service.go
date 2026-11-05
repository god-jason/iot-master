// Package broker 内嵌 MQTT Broker：设备连接鉴权、上行消息分发、设备注册与出站指令请求-响应。
//
// 主题统一为 $thing/{device_id}/{suffix}，主题常量与消息结构复用 internal/mqtt。
package broker

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/mqtt"
	"github.com/god-jason/iot-master/internal/ota"
	"github.com/god-jason/iot-master/internal/types"
	beeq "github.com/god-jason/iot-master/pkg/beeq"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"xorm.io/builder"
)

// 出站指令等待应答的超时时间
const (
	// defaultTimeout 属性读写、配置查询应答超时
	defaultTimeout = 10 * time.Second
	// controlTimeout 服务调用、配置下发、数据库同步应答超时
	controlTimeout = 30 * time.Second
)

// Service 内嵌 broker 的消息服务，实现 beeq.BrokerHandler。
// 负责设备连接鉴权、上行消息分发、设备注册与出站指令的请求-响应匹配。
type Service struct {
	server *beeq.Server          // 内嵌 broker 服务端实例，Attach 之后才可用于下发
	rm     *mqtt.ResponseManager // 出站指令的请求-响应匹配（按消息 id）
}

// NewService 创建内嵌 broker 消息服务
func NewService() *Service {
	return &Service{rm: mqtt.NewResponseManager()}
}

// Attach 绑定 broker 服务端实例（NewServer 之后调用）
func (s *Service) Attach(server *beeq.Server) {
	s.server = server
}

// ===== beeq.BrokerHandler =====

// OnAuth 设备连接鉴权：ClientID=设备ID，Username=物模型ID，
// Password=HMAC-MD5(ModelSecret, "{ModelID}/{DeviceID}/{ModelSecret}")。
// 鉴权通过后构造 Device 并注册到 iot registry。
func (s *Service) OnAuth(session *beeq.Session, password string) bool {
	deviceID := session.ClientID
	modelID := session.Username

	if deviceID == "" || modelID == "" {
		logger.Warn("设备鉴权失败：ClientID 或 Username 为空", "clientID", deviceID, "modelID", modelID)
		return false
	}

	var device types.Device
	found, err := db.FindOne("device", builder.Eq{"id": deviceID, "model_id": modelID}, &device)
	if err != nil || !found || device.ID == "" {
		logger.Warn("设备鉴权失败：设备不存在或不属于该物模型", "clientID", deviceID, "modelID", modelID, "error", err)
		return false
	}
	if device.Disabled {
		logger.Warn("设备鉴权失败：设备已被禁用", "clientID", deviceID)
		return false
	}

	var model types.Model
	found, err = db.FindOne("model", builder.Eq{"id": modelID}, &model)
	if err != nil || !found || model.ID == "" {
		logger.Warn("设备鉴权失败：物模型不存在", "clientID", deviceID, "modelID", modelID, "error", err)
		return false
	}

	content := fmt.Sprintf("%s/%s/%s", modelID, deviceID, model.Secret)
	h := hmac.New(md5.New, []byte(model.Secret))
	h.Write([]byte(content))
	if !hmac.Equal([]byte(password), []byte(hex.EncodeToString(h.Sum(nil)))) {
		logger.Warn("设备鉴权失败：密码验证失败", "clientID", deviceID, "modelID", modelID)
		return false
	}

	session.DeviceSN = device.SN
	session.ModelID = modelID

	s.registerDevice(&device)
	logger.Info("设备鉴权通过", "clientID", deviceID, "modelID", modelID, "sn", device.SN)
	return true
}

// OnMessage 上行消息分发：$thing/{device_id}/{suffix}
func (s *Service) OnMessage(session *beeq.Session, topic string, payload []byte, qos byte) {
	deviceID, suffix, ok := parseTopic(topic)
	if !ok {
		logger.Warn("消息 Topic 格式不正确", "topic", topic)
		return
	}
	if deviceID != session.ClientID {
		logger.Warn("消息 Topic 与连接设备不一致", "topic", topic, "clientID", session.ClientID)
		return
	}

	// 按后缀分发：*Report 走业务处理，*Resp 回投给等待中的出站请求
	switch suffix {
	case PropertyReport:
		s.handlePropertyReport(deviceID, payload)
	case EventReport:
		s.handleEventReport(deviceID, payload)
	case StatusReport:
		s.handleStatusReport(deviceID, payload)
	case OTAReport:
		s.handleOTAReport(deviceID, payload)
	case PropertyReportResp, PropertySetResp, PropertyGetResp, ServiceCallResp, OTAReportResp,
		SettingSetResp, SettingGetResp, DatabaseSyncResp, DatabaseLoadResp:
		s.handleResponse(deviceID, suffix, payload)
	default:
		logger.Warn("未知消息类型", "topic", topic, "deviceID", deviceID)
	}
}

// OnDisconnect 设备正常断开
func (s *Service) OnDisconnect(session *beeq.Session) {
	logger.Info("设备断开连接", "clientID", session.ClientID)
	session.SetOnline(false)
	if dev, ok := iot.GetDevice(session.ClientID); ok {
		dev.SetOffline()
	}
}

// OnLose 设备连接丢失
func (s *Service) OnLose(session *beeq.Session, err error) {
	logger.Warn("设备连接丢失", "clientID", session.ClientID, "error", err)
	session.SetOnline(false)
	if dev, ok := iot.GetDevice(session.ClientID); ok {
		dev.SetOffline()
	}
}

// OnError 设备错误
func (s *Service) OnError(session *beeq.Session, err error) {
	if session != nil {
		logger.Error("设备错误", "clientID", session.ClientID, "error", err)
		return
	}
	logger.Error("设备错误", "error", err)
}

// ===== 上行消息处理 =====

// handlePropertyReport 属性上报
func (s *Service) handlePropertyReport(deviceID string, payload []byte) {
	var m PropertyReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析属性上报失败", "deviceID", deviceID, "error", err)
		s.reply(deviceID, PropertyReportResp, m.ID, 0, "invalid message format")
		return
	}

	dev := s.deviceOf(deviceID)
	if dev == nil {
		s.reply(deviceID, PropertyReportResp, m.ID, 0, "device not found")
		return
	}
	dev.SetOnline()
	if len(m.Params) > 0 {
		// 按实际处理结果返回 ret：失败时告知设备，避免永远回 success
		if err := dev.ReportProperties(m.Params, msgTime(m.Time)); err != nil {
			logger.Warn("属性上报处理失败", "deviceID", deviceID, "id", m.ID, "error", err)
			s.reply(deviceID, PropertyReportResp, m.ID, 0, "property report failed")
			return
		}
	}
	s.reply(deviceID, PropertyReportResp, m.ID, 1, "success")
}

// handleEventReport 事件上报
func (s *Service) handleEventReport(deviceID string, payload []byte) {
	var m EventReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析事件上报失败", "deviceID", deviceID, "error", err)
		s.reply(deviceID, EventReportResp, m.ID, 0, "invalid message format")
		return
	}

	dev := s.deviceOf(deviceID)
	if dev == nil {
		s.reply(deviceID, EventReportResp, m.ID, 0, "device not found")
		return
	}
	dev.SetOnline()
	if m.Event.Name == "" {
		s.reply(deviceID, EventReportResp, m.ID, 0, "event name required")
		return
	}
	_ = dev.ReportEvent(m.Event.Name, m.Event.Params, time.Now())
	s.reply(deviceID, EventReportResp, m.ID, 1, "success")
}

// handleStatusReport 状态上报
func (s *Service) handleStatusReport(deviceID string, payload []byte) {
	var m StatusReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析状态上报失败", "deviceID", deviceID, "error", err)
		s.reply(deviceID, StatusReportResp, m.ID, 0, "invalid message format")
		return
	}

	dev := s.deviceOf(deviceID)
	if dev == nil {
		s.reply(deviceID, StatusReportResp, m.ID, 0, "device not found")
		return
	}
	dev.SetOnline()
	_ = dev.ReportStatus(m.Params, msgTime(m.Time))
	s.reply(deviceID, StatusReportResp, m.ID, 1, "success")
}

// handleOTAReport OTA 升级进度上报
func (s *Service) handleOTAReport(deviceID string, payload []byte) {
	var m OTAReportMsg
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Error("解析 OTA 上报失败", "deviceID", deviceID, "error", err)
		s.reply(deviceID, OTAReportResp, m.ID, 0, "invalid message format")
		return
	}
	if m.TaskID == "" {
		s.reply(deviceID, OTAReportResp, m.ID, 0, "task_id required")
		return
	}

	dev := s.deviceOf(deviceID)
	if dev == nil {
		s.reply(deviceID, OTAReportResp, m.ID, 0, "device not found")
		return
	}
	dev.SetOnline()
	ota.HandleReport(deviceID, m.TaskID, m.Stage, m.Progress, m.Error, m.Version)
	s.reply(deviceID, OTAReportResp, m.ID, 1, "success")
}

// handleResponse 设备应答，按消息 id 回投给等待中的出站请求
func (s *Service) handleResponse(deviceID, suffix string, payload []byte) {
	var m ResponseMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		logger.Warn("解析设备应答失败", "deviceID", deviceID, "suffix", suffix, "error", err)
		return
	}
	if s.rm.HandleResponse(m.ID, &m) {
		return
	}
	logger.Debug("收到未匹配的设备应答", "deviceID", deviceID, "suffix", suffix, "id", m.ID, "ret", m.Ret)
}

// ===== 出站指令 =====

// request 下发指令到 suffix 主题，按 id 等待设备应答；超时返回 ErrResponseTimeout
func (s *Service) request(deviceID, suffix string, payload any, msgID string, timeout time.Duration) (*ResponseMessage, error) {
	if s.server == nil {
		return nil, beeq.ErrNoBroker
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	ch, err := s.rm.Add(msgID)
	if err != nil {
		return nil, err
	}

	if !s.publish(deviceID, suffix, data) {
		s.rm.Remove(msgID)
		return nil, beeq.ErrPublishFailed
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		s.rm.Remove(msgID)
		return nil, mqtt.ErrResponseTimeout
	}
}

// publish 向设备发送消息
func (s *Service) publish(deviceID, suffix string, data []byte) bool {
	if s.server == nil {
		logger.Error("内嵌 broker 未初始化", "deviceID", deviceID, "suffix", suffix)
		return false
	}
	return s.server.Publish(deviceID, BuildTopic(deviceID, suffix), data, 1)
}

// reply 向设备回复上报应答
func (s *Service) reply(deviceID, suffix, msgID string, ret int, message string) {
	if msgID == "" {
		return
	}
	data, err := json.Marshal(ReportResponseMessage{ID: msgID, Ret: ret, Message: message})
	if err != nil {
		logger.Error("序列化上报应答失败", "deviceID", deviceID, "error", err)
		return
	}
	s.publish(deviceID, suffix, data)
}

// ===== 设备注册 =====

// registerDevice 构造 Device 并注册到 iot registry（已注册则标记在线）
func (s *Service) registerDevice(dbDev *types.Device) iot.Device {
	if dev, ok := iot.GetDevice(dbDev.ID); ok {
		dev.SetOnline()
		return dev
	}
	dev := NewDevice(dbDev, s)
	dev.SetOnline()
	iot.Register(dbDev.ID, dev)
	logger.Info("内嵌 broker 设备已注册", "device", dbDev.ID)
	return dev
}

// deviceOf 从 registry 取设备；未注册则查 DB 构造并注册
func (s *Service) deviceOf(deviceID string) iot.Device {
	if dev, ok := iot.GetDevice(deviceID); ok {
		return dev
	}

	var dbDev types.Device
	found, err := db.FindOne("device", builder.Eq{"id": deviceID}, &dbDev)
	if err != nil {
		logger.Warn("查 DB 注册设备失败", "device", deviceID, "error", err)
		return nil
	}
	if !found || dbDev.ID == "" {
		logger.Debug("设备未在 DB 注册，跳过 registry", "device", deviceID)
		return nil
	}
	return s.registerDevice(&dbDev)
}

// ===== 公共辅助 =====

// parseTopic 从 $thing/{device_id}/{suffix} 解析设备 ID 与后缀
func parseTopic(topic string) (deviceID, suffix string, ok bool) {
	parts := strings.SplitN(topic, "/", 3)
	if len(parts) != 3 || parts[0] != TopicPrefix || parts[1] == "" || parts[2] == "" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// msgTime 将毫秒级时间戳转为 time.Time，缺省用当前时间
func msgTime(ms int64) time.Time {
	if ms > 0 {
		return time.UnixMilli(ms)
	}
	return time.Now()
}
