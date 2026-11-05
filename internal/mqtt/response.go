package mqtt

import (
	"errors"
	"sync"
)

// 应答匹配相关错误
var (
	// ErrResponseTimeout 等待设备应答超时
	ErrResponseTimeout = errors.New("mqtt: response timeout")
	// ErrDuplicateRequest 同一消息 id 已有待匹配请求
	ErrDuplicateRequest = errors.New("mqtt: request id already exists")
)

// responseManager 全局请求-响应匹配管理器（按消息 id 关联下发指令与设备应答）
var responseManager = NewResponseManager()

// ResponseManager 请求-响应匹配管理器
type ResponseManager struct {
	mu      sync.RWMutex                     // 保护 pending 的并发读写
	pending map[string]chan *ResponseMessage // 消息 id → 等待应答的通道
}

// NewResponseManager 创建响应管理器
func NewResponseManager() *ResponseManager {
	return &ResponseManager{pending: make(map[string]chan *ResponseMessage)}
}

// Add 注册一个待匹配请求，返回接收应答的通道
func (rm *ResponseManager) Add(msgID string) (chan *ResponseMessage, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if _, exists := rm.pending[msgID]; exists {
		return nil, ErrDuplicateRequest
	}
	ch := make(chan *ResponseMessage, 1) // 缓冲 1，保证应答投递不会被阻塞
	rm.pending[msgID] = ch
	return ch, nil
}

// Remove 移除待匹配请求
func (rm *ResponseManager) Remove(msgID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.pending, msgID)
}

// HandleResponse 按 id 投递设备应答，返回是否匹配到等待中的请求
func (rm *ResponseManager) HandleResponse(msgID string, resp *ResponseMessage) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	ch, ok := rm.pending[msgID]
	if !ok {
		return false
	}
	// 非阻塞投递，避免等待方已超时放弃时阻塞分发协程
	select {
	case ch <- resp:
	default:
	}
	delete(rm.pending, msgID)
	return true
}
