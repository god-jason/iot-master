package beeq

import "errors"

var (
	ErrAuthFailed        = errors.New("authentication failed")
	ErrProtocol          = errors.New("invalid protocol")
	ErrClientID          = errors.New("empty clientID")
	ErrPacket            = errors.New("invalid packet")
	ErrNoBroker          = errors.New("broker not initialized")
	ErrNoResponseManager = errors.New("response manager not initialized")
	ErrInvalidMessageID  = errors.New("invalid message ID")
	ErrPublishFailed     = errors.New("publish failed")
)

type BrokerHandler interface {
	OnAuth(session *Session, password string) bool
	OnMessage(session *Session, topic string, payload []byte, qos byte)
	OnDisconnect(session *Session)
	OnLose(session *Session, err error)
	OnError(session *Session, err error)
}

type DefaultHandler struct{}

func (h *DefaultHandler) OnAuth(session *Session, password string) bool {
	return true
}

func (h *DefaultHandler) OnMessage(session *Session, topic string, payload []byte, qos byte) {
}

func (h *DefaultHandler) OnDisconnect(session *Session) {
}

func (h *DefaultHandler) OnLose(session *Session, err error) {
}

func (h *DefaultHandler) OnError(session *Session, err error) {
}
