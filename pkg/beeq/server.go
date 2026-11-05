package beeq

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Server struct {
	gnet.BuiltinEventEngine
	eng           gnet.Engine   // gnet引擎
	sessions      sync.Map      // 客户端会话映射：deviceID -> Session
	blacklist     *Blacklist    // 自动黑名单管理器
	nextSessionID uint64        // 下一个会话ID
	nextPacketID  uint64        // 下一个消息包ID
	handler       BrokerHandler // 事件处理器
	running       bool          // 运行状态
	opts          Options       // 服务器配置选项
}

func NewServer(opts Options, handler BrokerHandler) *Server {
	if opts.MaxPacketSize == 0 {
		opts.MaxPacketSize = DefaultOptions().MaxPacketSize
	}
	if opts.MaxConnPending == 0 {
		opts.MaxConnPending = DefaultOptions().MaxConnPending
	}
	if opts.SupportedProtocols == nil {
		opts.SupportedProtocols = DefaultOptions().SupportedProtocols
	}
	if opts.BlacklistMaxAttempts == 0 {
		opts.BlacklistMaxAttempts = DefaultOptions().BlacklistMaxAttempts
	}
	if opts.BlacklistBlockTime == 0 {
		opts.BlacklistBlockTime = DefaultOptions().BlacklistBlockTime
	}
	if opts.KeepAliveTimeout == 0 {
		opts.KeepAliveTimeout = DefaultOptions().KeepAliveTimeout
	}

	if handler == nil {
		handler = &DefaultHandler{}
	}

	return &Server{
		blacklist: NewBlacklist(opts.BlacklistMaxAttempts, opts.BlacklistBlockTime),
		handler:   handler,
		running:   true,
		opts:      opts,
	}
}

func (s *Server) Listen(addr string) error {
	return gnet.Run(s, addr, s.opts.toGnetOptions()...)
}

func (s *Server) OnBoot(eng gnet.Engine) gnet.Action {
	s.eng = eng
	return gnet.None
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	ip := c.RemoteAddr().String()
	if s.blacklist.IsBlocked(ip) {
		return nil, gnet.Close
	}
	return nil, gnet.None
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	ctx := c.Context()
	if ctx == nil {
		return s.handleHandshake(c)
	}
	session := ctx.(*Session)

	session.Active()

	buf, _ := c.Peek(c.InboundBuffered())
	if len(buf) == 0 {
		return gnet.None
	}

	if len(buf) > s.opts.MaxPacketSize {
		return gnet.Close
	}

	totalConsumed := 0
	for len(buf) > 0 {
		packet, consumed := ReadPacket(buf)
		if packet == nil {
			break
		}

		totalConsumed += consumed
		buf = buf[consumed:]
		s.handlePacket(session, packet, c)
	}

	if totalConsumed > 0 {
		c.Discard(totalConsumed)
	}

	return gnet.None
}

func (s *Server) handleHandshake(c gnet.Conn) gnet.Action {
	buf, _ := c.Peek(c.InboundBuffered())
	if len(buf) == 0 {
		return gnet.None
	}

	packet, consumed := ReadPacket(buf)
	if packet == nil || packet.Type() != CMD_CONNECT {
		return gnet.Close
	}

	c.Discard(consumed)
	buf = buf[consumed:]

	cp := ParseConnect(packet.Payload)
	if cp == nil {
		return gnet.Close
	}

	if !s.isProtocolSupported(cp.ProtocolVersion) {
		c.Write(BuildConnack(false, 0x01).Build())
		return gnet.Close
	}

	if cp.ClientID == "" {
		c.Write(BuildConnack(false, 0x02).Build())
		return gnet.Close
	}

	session := NewSession(
		atomic.AddUint64(&s.nextSessionID, 1),
		c,
		cp.ClientID,
		cp.Username,
		cp.CleanSession,
		cp.KeepAlive,
		s.opts.MaxConnPending,
	)

	if !s.handler.OnAuth(session, cp.Password) {
		ip := c.RemoteAddr().String()
		s.blacklist.Record(ip)
		c.Write(BuildConnack(false, 0x05).Build())
		s.handler.OnError(session, ErrAuthFailed)
		return gnet.Close
	}

	c.SetContext(session)
	s.sessions.Store(cp.ClientID, session)

	c.Write(BuildConnack(false, 0x00).Build())

	totalConsumed := 0
	for len(buf) > 0 {
		packet, consumed := ReadPacket(buf)
		if packet == nil {
			break
		}
		totalConsumed += consumed
		buf = buf[consumed:]
		s.handlePacket(session, packet, c)
	}

	if totalConsumed > 0 {
		c.Discard(totalConsumed)
	}

	return gnet.None
}

func (s *Server) handlePacket(session *Session, packet *Packet, c gnet.Conn) {
	switch packet.Type() {
	case CMD_PUBLISH:
		s.handlePublish(session, packet, c)
	case CMD_PUBACK:
	case CMD_PUBREC:
		if len(packet.Payload) < 2 {
			return
		}
		packetID := uint16(packet.Payload[0])<<8 | uint16(packet.Payload[1])
		c.Write(BuildPubrel(packetID).Build())
	case CMD_PUBREL:
		if len(packet.Payload) < 2 {
			return
		}
		packetID := uint16(packet.Payload[0])<<8 | uint16(packet.Payload[1])
		c.Write(BuildPubcomp(packetID).Build())
	case CMD_PUBCOMP:
	case CMD_SUBSCRIBE:
		sp := ParseSubscribe(packet.Payload)
		if sp == nil {
			return
		}
		qosList := make([]byte, 0, len(sp.Topics))
		for _, topic := range sp.Topics {
			qosList = append(qosList, topic.QoS)
		}
		c.Write(BuildSuback(sp.PacketID, qosList).Build())
	case CMD_UNSUBSCRIBE:
		up := ParseUnsubscribe(packet.Payload)
		if up == nil {
			return
		}
		c.Write(BuildUnsuback(up.PacketID).Build())
	case CMD_PINGREQ:
		c.Write(BuildPingresp().Build())
	case CMD_DISCONNECT:
		session.Close()
	}
}

func (s *Server) handlePublish(session *Session, packet *Packet, c gnet.Conn) {
	pp := ParsePublish(packet.Payload, packet.QoS())
	if pp == nil {
		return
	}
	pp.Dup = packet.Dup()
	pp.QoS = packet.QoS()
	pp.Retain = packet.Retain()

	if pp.QoS == 1 {
		c.Write(BuildPuback(pp.PacketID).Build())
	} else if pp.QoS == 2 {
		c.Write(BuildPubrec(pp.PacketID).Build())
		session.AddPubRel(pp.PacketID, pp)
	}

	s.handler.OnMessage(session, pp.TopicName, pp.Payload, pp.QoS)
}

func (s *Server) Publish(clientID string, topic string, payload []byte, qos byte) bool {
	val, ok := s.sessions.Load(clientID)
	if !ok {
		return false
	}

	session := val.(*Session)
	if session.IsClosed() {
		return false
	}

	pp := &PublishPacket{
		TopicName: topic,
		PacketID:  uint16(atomic.AddUint64(&s.nextPacketID, 1) & 0xFFFF),
		Payload:   payload,
		Dup:       false,
		QoS:       qos,
		Retain:    false,
	}

	data := BuildPublish(pp).Build()
	if !session.CanWrite(len(data)) {
		return false
	}

	deviceID := session.ClientID
	err := session.Conn.AsyncWrite(data, func(c gnet.Conn, err error) error {
		if err != nil {
			session.Close()
			s.sessions.Delete(deviceID)
		}
		return nil
	})

	if err != nil {
		session.Close()
		s.sessions.Delete(deviceID)
		return false
	}

	return true
}

func (s *Server) OnClose(c gnet.Conn, err error) gnet.Action {
	ctx := c.Context()
	if ctx == nil {
		return gnet.None
	}
	session := ctx.(*Session)
	if session != nil {
		if s.opts.CleanSession {
			s.sessions.Delete(session.ClientID)
		}
		session.Close()
		session.SetOnline(false)
		if err != nil {
			s.handler.OnLose(session, err)
		} else {
			s.handler.OnDisconnect(session)
		}
	}
	return gnet.None
}

func (s *Server) OnTick() (time.Duration, gnet.Action) {
	s.blacklist.Cleanup()
	s.cleanupInactive()
	return 30 * time.Second, gnet.None
}

func (s *Server) cleanupInactive() {
	now := time.Now().Unix()
	s.sessions.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		if session.KeepAlive > 0 && now-session.LastActive > int64(session.KeepAlive)*2 {
			session.Close()
		}
		return true
	})
}

func (s *Server) Close() {
	s.running = false
	s.sessions.Range(func(key, value interface{}) bool {
		session := value.(*Session)
		session.Close()
		return true
	})
	_ = s.eng.Stop(context.Background())
}

func (s *Server) GetSessionCount() int {
	count := 0
	s.sessions.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

func (s *Server) GetSession(clientID string) (*Session, bool) {
	val, ok := s.sessions.Load(clientID)
	if !ok {
		return nil, false
	}
	session := val.(*Session)
	// if session.IsClosed() {
	// 	return nil, false
	// }
	return session, true
}

func (s *Server) isProtocolSupported(version byte) bool {
	for _, v := range s.opts.SupportedProtocols {
		if v == version {
			return true
		}
	}
	return false
}
