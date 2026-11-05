package beeq

import (
	"sync"
	"time"
)

// Blacklist 自动黑名单管理器，基于IP地址进行连接频率限制
// 用于防止恶意客户端频繁连接攻击
type Blacklist struct {
	mu        sync.RWMutex               // 读写锁，保护并发访问
	entries   map[string]*blacklistEntry // IP地址到记录的映射
	threshold int                        // 触发封禁的失败次数阈值
	window    time.Duration              // 统计窗口时间
}

// blacklistEntry 单个IP的黑名单记录
type blacklistEntry struct {
	count     int       // 失败次数
	firstTime time.Time // 首次失败时间
	lastTime  time.Time // 最近失败时间
}

// NewBlacklist 创建新的黑名单管理器
// threshold: 触发封禁的失败次数阈值
// window: 统计窗口时间，超过此时间后重置计数
func NewBlacklist(threshold int, window time.Duration) *Blacklist {
	return &Blacklist{
		entries:   make(map[string]*blacklistEntry),
		threshold: threshold,
		window:    window,
	}
}

// Record 记录一次连接失败
// ip: 客户端IP地址
// 返回值: true表示已被封禁，false表示未被封禁
func (bl *Blacklist) Record(ip string) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	entry, ok := bl.entries[ip]
	if !ok {
		// 首次失败，创建新记录
		entry = &blacklistEntry{
			count:     1,
			firstTime: now,
			lastTime:  now,
		}
		bl.entries[ip] = entry
		return false
	}

	// 如果超过统计窗口，重置计数
	if now.Sub(entry.firstTime) > bl.window {
		entry.count = 1
		entry.firstTime = now
		entry.lastTime = now
		return false
	}

	// 增加失败计数
	entry.count++
	entry.lastTime = now

	// 判断是否达到封禁阈值
	return entry.count >= bl.threshold
}

// IsBlocked 检查IP是否被封禁
// ip: 客户端IP地址
// 返回值: true表示已被封禁，false表示未被封禁
func (bl *Blacklist) IsBlocked(ip string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	entry, ok := bl.entries[ip]
	if !ok {
		return false
	}

	// 如果封禁时间超过2倍窗口时间，自动解封
	if time.Now().Sub(entry.lastTime) > bl.window*2 {
		delete(bl.entries, ip)
		return false
	}

	return entry.count >= bl.threshold
}

// Unblock 手动解封IP地址
// ip: 要解封的IP地址
func (bl *Blacklist) Unblock(ip string) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	delete(bl.entries, ip)
}

// Cleanup 清理过期的黑名单记录
// 定期调用此方法清理长时间未活动的IP记录
func (bl *Blacklist) Cleanup() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	for ip, entry := range bl.entries {
		if now.Sub(entry.lastTime) > bl.window*2 {
			delete(bl.entries, ip)
		}
	}
}
