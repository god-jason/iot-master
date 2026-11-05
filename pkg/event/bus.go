package event

import (
	"strings"
	"sync"

	"github.com/god-jason/iot-master/pkg/pool"
)

// =====================================================================
// eventBus — 通配符，订阅树（trie）索引
//
// 关键设计：
//   - 每个节点：children map[segment]*node + handlers []Handler
//   - 通配符段 "+" 走 children["+"]
//   - 通配符段 "#" 只可能出现在 filter 最后一级，挂在 children["#"] 的 handlers 上
//   - Publish 时按 topic 段 DFS，遇到 "+" 继续匹配子节点，遇到 "#" 直接取该节点 handlers
//   - 时间复杂度 O(len(topic) + 扇出)，与订阅总数无关
// =====================================================================

const (
	wildcardSingle = "+" // 单级通配
	wildcardMulti  = "#" // 多级通配（必须是 filter 最后一段）
)

type node struct {
	handlers []Handler        // 到达此节点的 filter 的 handler
	children map[string]*node // 下一级 segment → child node
	hasHash  bool             // 该节点下是否挂了 # 通配的 handlers（缓存）
}

func newNode() *node {
	return &node{children: make(map[string]*node)}
}

// eventBus 标准事件总线，支持通配符（+ / #）
type eventBus struct {
	mu   sync.RWMutex
	root *node
}

// NewEventBus 创建通配符事件总线
func NewEventBus() *eventBus {
	return &eventBus{root: newNode()}
}

// Subscribe 订阅 filter，返回取消订阅的函数（幂等）。
// filter 支持：
//   - "a/b/c"     精确匹配
//   - "a/+/c"     单级通配
//   - "a/#"       多级通配（必须最后一段）
func (b *eventBus) Subscribe(filter string, handler Handler) func() {
	segments := splitAndValidate(filter)
	b.mu.Lock()
	b.insert(segments, handler)
	b.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() { b.Unsubscribe(filter, handler) })
	}
}

// Unsubscribe 显式取消订阅（按 filter + handler）
func (b *eventBus) Unsubscribe(filter string, handler Handler) {
	segments := splitAndValidate(filter)
	b.mu.Lock()
	defer b.mu.Unlock()

	target := b.findNode(segments)
	if target == nil {
		return
	}

	key := handlerKey(handler)
	for i, h := range target.handlers {
		if handlerKey(h) == key {
			target.handlers = append(target.handlers[:i], target.handlers[i+1:]...)
			break
		}
	}
	if len(target.handlers) == 0 {
		b.prune(segments)
	}
}

func (b *eventBus) findNode(segments []string) *node {
	cur := b.root
	for _, seg := range segments {
		child, ok := cur.children[seg]
		if !ok {
			return nil
		}
		cur = child
	}
	return cur
}

// Publish 异步发布，所有匹配 filter 的 handler 通过 pool 执行
func (b *eventBus) Publish(topic string, payload any) {
	b.mu.RLock()
	handlers := b.collect(topic)
	b.mu.RUnlock()

	for _, h := range handlers {
		h := h
		pool.Submit(func() { safeCall(h, topic, payload) })
	}
}

// Reset 清空全部订阅
func (b *eventBus) Reset() {
	b.mu.Lock()
	b.root = newNode()
	b.mu.Unlock()
}

// Len 订阅总数（粗粒度：所有 handler 数量）
func (b *eventBus) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return countHandlers(b.root)
}

func countHandlers(n *node) int {
	if n == nil {
		return 0
	}
	total := len(n.handlers)
	for _, c := range n.children {
		total += countHandlers(c)
	}
	return total
}

// ===== trie 操作 =====

func (b *eventBus) insert(segments []string, handler Handler) *node {
	cur := b.root
	for _, seg := range segments {
		child, ok := cur.children[seg]
		if !ok {
			child = newNode()
			cur.children[seg] = child
			if seg == wildcardMulti {
				cur.hasHash = true
			}
		}
		cur = child
	}
	cur.handlers = append(cur.handlers, handler)
	return cur
}

func (b *eventBus) prune(segments []string) {
	path := make([]*node, 0, len(segments)+1)
	path = append(path, b.root)
	cur := b.root
	for _, seg := range segments {
		child, ok := cur.children[seg]
		if !ok {
			return
		}
		path = append(path, child)
		cur = child
	}
	for i := len(path) - 1; i > 0; i-- {
		n := path[i]
		parent := path[i-1]
		seg := segments[i-1]
		if len(n.handlers) == 0 && len(n.children) == 0 {
			delete(parent.children, seg)
			if seg == wildcardMulti {
				parent.hasHash = false
			}
		} else {
			break
		}
	}
}

func (b *eventBus) collect(topic string) []Handler {
	parts := strings.Split(topic, "/")
	result := make([]Handler, 0, 8)

	var dfs func(n *node, idx int)
	dfs = func(n *node, idx int) {
		if n == nil {
			return
		}
		if idx >= len(parts) {
			result = append(result, n.handlers...)
			return
		}
		if hashChild, ok := n.children[wildcardMulti]; ok {
			result = append(result, hashChild.handlers...)
		}
		if child, ok := n.children[parts[idx]]; ok {
			dfs(child, idx+1)
		}
		if child, ok := n.children[wildcardSingle]; ok {
			dfs(child, idx+1)
		}
	}

	dfs(b.root, 0)
	return result
}

func splitAndValidate(filter string) []string {
	if filter == "" {
		return nil
	}
	segments := strings.Split(filter, "/")
	for i, s := range segments {
		if s == wildcardMulti && i != len(segments)-1 {
			segments = segments[:i+1]
			break
		}
	}
	return segments
}
