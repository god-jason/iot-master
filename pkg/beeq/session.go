package beeq

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"
)

type Session struct {
	ID         uint64                    // 会话唯一标识
	Conn       gnet.Conn                 // gnet连接
	ClientID   string                    // 客户端ID，即设备ID
	DeviceSN   string                    // 设备SN
	Username   string                    // 用户名，即物模型ID
	ModelID    string                    // 物模型ID
	Clean      bool                      // 是否清除会话标志
	KeepAlive  uint16                    // 保持连接时间(秒)
	LastActive int64                     // 最后活跃时间戳
	PubRel     map[uint16]*PublishPacket // QoS 2消息暂存
	mu         sync.RWMutex              // 读写锁，保护并发访问
	closed     int32                     // 关闭标志：1表示已关闭，0表示正常
	handshaked bool                      // 是否已完成MQTT握手
	maxPending int                       // 最大待发送字节数
	Online     bool                      // 在线状态
}

func NewSession(id uint64, conn gnet.Conn, clientID, username string, clean bool, keepAlive uint16, maxPending int) *Session {
	if maxPending <= 0 {
		maxPending = 1024 * 1024
	}
	return &Session{
		ID:         id,
		Conn:       conn,
		ClientID:   clientID,
		Username:   username,
		Clean:      clean,
		KeepAlive:  keepAlive,
		LastActive: time.Now().Unix(),
		PubRel:     make(map[uint16]*PublishPacket),
		handshaked: true,
		maxPending: maxPending,
		Online:     true,
	}
}

func (s *Session) CanWrite(size int) bool {
	return s.Conn.OutboundBuffered()+size <= s.maxPending
}

func (s *Session) Active() {
	s.LastActive = time.Now().Unix()
}

func (s *Session) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		s.Conn.Close()
	}
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}

func (s *Session) AddPubRel(packetID uint16, pp *PublishPacket) {
	s.mu.Lock()
	s.PubRel[packetID] = pp
	s.mu.Unlock()
}

func (s *Session) GetPubRel(packetID uint16) (*PublishPacket, bool) {
	s.mu.RLock()
	pp, ok := s.PubRel[packetID]
	s.mu.RUnlock()
	return pp, ok
}

func (s *Session) RemovePubRel(packetID uint16) {
	s.mu.Lock()
	delete(s.PubRel, packetID)
	s.mu.Unlock()
}

func (s *Session) SetOnline(online bool) {
	s.Online = online
}

func (s *Session) IsOnline() bool {
	return s.Online
}
