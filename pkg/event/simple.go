// Package event — 进程内事件总线（pub/sub）。
//
// 提供两种实现（均私有类型，通过公开构造函数创建）：
//
//  1. **simpleBus** — 精确 topic 匹配，map[string][]Handler 索引。
//     适合 topic 数量少、不需要通配符的场景。
//
//  2. **eventBus** — MQTT 风格通配符（`+` 单级 / `#` 多级），
//     内部用订阅树（trie）索引，Publish 时间复杂度 O(深度 + 扇出)，
//     不随订阅总数增长。适合高 topic 基数、大量订阅者的核心链路，
//     其他库可以直接复用。
//
// 两者共享 Handler 签名，Subscribe/Unsubscribe/Publish/PublishSync/Reset 方法完全一致。
package event

import (
	"sync"

	"github.com/god-jason/iot-master/pkg/pool"
)

// Handler 事件回调函数
type Handler func(topic string, payload any)

// =====================================================================
// simpleBus — 精确 topic 匹配，map 索引
// =====================================================================

// simpleBus 简单事件总线（仅精确匹配，无通配符）
type simpleBus struct {
	mu    sync.RWMutex
	index map[string][]Handler // topic → handlers
}

// NewSimple 创建精确匹配事件总线
func NewSimple() *simpleBus {
	return &simpleBus{index: make(map[string][]Handler)}
}

// Subscribe 订阅精确 topic，返回取消订阅的函数（幂等）
func (b *simpleBus) Subscribe(topic string, handler Handler) func() {
	b.mu.Lock()
	b.index[topic] = append(b.index[topic], handler)
	b.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() { b.Unsubscribe(topic, handler) })
	}
}

// Unsubscribe 显式取消订阅（与 Subscribe 返回的取消函数等价）
func (b *simpleBus) Unsubscribe(topic string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	list, ok := b.index[topic]
	if !ok {
		return
	}
	key := handlerKey(handler)
	for i, h := range list {
		if handlerKey(h) == key {
			b.index[topic] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(b.index[topic]) == 0 {
		delete(b.index, topic)
	}
}

// Publish 异步发布，匹配的 handler 通过 pool 执行
func (b *simpleBus) Publish(topic string, payload any) {
	b.mu.RLock()
	list := append([]Handler(nil), b.index[topic]...)
	b.mu.RUnlock()

	for _, h := range list {
		h := h
		pool.Submit(func() { safeCall(h, topic, payload) })
	}
}

// Reset 清空所有订阅
func (b *simpleBus) Reset() {
	b.mu.Lock()
	b.index = make(map[string][]Handler)
	b.mu.Unlock()
}
