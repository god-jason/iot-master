package cache

import (
	"sync"
	"time"
)

// expiredItem 带 TTL 的缓存项：存值 / 上次加载错误（error != nil 表示曾加载失败，
// 到期后会自动清除让下次 Get 重试）
type expiredItem[V any] struct {
	value    V
	err      error
	expireAt time.Time
	loadTime time.Time
}

// ExpiredCache 泛型带 TTL 缓存
//
//   - 快路径：RWMutex 读锁命中 → 检查 expireAt（过期算 miss）
//   - 慢路径：singleflight，同 key 并发请求共享一次 loader 调用
//   - 错误缓存：loader 返回 err 时也按 TTL 入缓存，到期自动清除后重试，避免 DB 风暴
//
// 设计参考 Cache（singleflight + 错误缓存）+ 保留 TTL 过期语义。
type ExpiredCache[K comparable, V any] struct {
	data   map[K]*expiredItem[V]
	loader LoaderFunc[K, V]
	ttl    time.Duration

	mu       sync.RWMutex          // 保护 data
	flight   sync.Mutex            // 保护 inflight
	inflight map[K]*sync.WaitGroup // singleflight：同 key 并发请求共享一次 loader
}

// NewExpiredCache 创建带 TTL 的缓存实例
func NewExpiredCache[K comparable, V any](loader LoaderFunc[K, V], ttl time.Duration) *ExpiredCache[K, V] {
	return &ExpiredCache[K, V]{
		data:     make(map[K]*expiredItem[V]),
		loader:   loader,
		ttl:      ttl,
		inflight: make(map[K]*sync.WaitGroup),
	}
}

// Get 获取缓存值
//
//	hit + 未过期 + err=nil     → 直接返回
//	hit + 未过期 + err!=nil    → 返回上次错误（TTL 内不重试）
//	hit + 已过期               → 走 singleflight 重新加载
//	miss                       → 走 singleflight 加载
func (c *ExpiredCache[K, V]) Get(key K) (V, error) {
	// 1. 快路径：读锁命中 + 未过期
	now := time.Now()
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()
	if ok && now.Before(item.expireAt) {
		var zero V
		if item.err != nil {
			return zero, item.err
		}
		return item.value, nil
	}

	// 2. 慢路径：singleflight（过期和 miss 都走这里）
	c.flight.Lock()
	wg, existing := c.inflight[key]
	if !existing {
		wg = &sync.WaitGroup{}
		wg.Add(1)
		c.inflight[key] = wg
	}
	c.flight.Unlock()

	if existing {
		wg.Wait()
		c.mu.RLock()
		item, ok = c.data[key]
		c.mu.RUnlock()
		if ok && time.Now().Before(item.expireAt) {
			var zero V
			if item.err != nil {
				return zero, item.err
			}
			return item.value, nil
		}
		var zero V
		return zero, ErrLoaderNotSet
	}

	// 我来加载
	defer func() {
		c.flight.Lock()
		delete(c.inflight, key)
		c.flight.Unlock()
		wg.Done()
	}()

	if c.loader == nil {
		var zero V
		return zero, ErrLoaderNotSet
	}

	v, err := c.loader(key)
	now = time.Now()
	c.mu.Lock()
	c.data[key] = &expiredItem[V]{
		value:    v,
		err:      err,
		expireAt: now.Add(c.ttl),
		loadTime: now,
	}
	c.mu.Unlock()
	return v, err
}

// Set 手动设置缓存值（同时清掉 err 标记）
func (c *ExpiredCache[K, V]) Set(key K, value V) {
	now := time.Now()
	c.mu.Lock()
	c.data[key] = &expiredItem[V]{
		value:    value,
		expireAt: now.Add(c.ttl),
		loadTime: now,
	}
	c.mu.Unlock()
}

// SetWithTTL 手动设置缓存值，使用自定义 TTL
func (c *ExpiredCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	now := time.Now()
	c.mu.Lock()
	c.data[key] = &expiredItem[V]{
		value:    value,
		expireAt: now.Add(ttl),
		loadTime: now,
	}
	c.mu.Unlock()
}

// SetErr 标记该 key 加载失败（后续 Get 在 TTL 内直接返回该 err，TTL 过后自动清除重试）
func (c *ExpiredCache[K, V]) SetErr(key K, err error) {
	now := time.Now()
	var zero V
	c.mu.Lock()
	c.data[key] = &expiredItem[V]{
		value:    zero,
		err:      err,
		expireAt: now.Add(c.ttl),
		loadTime: now,
	}
	c.mu.Unlock()
}

// Delete 删除缓存项（包括过期项和错误标记）；下次 Get 会走 loader
func (c *ExpiredCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// Refresh 强制重新加载（忽略旧值、错误标记和过期状态）
func (c *ExpiredCache[K, V]) Refresh(key K) (V, error) {
	c.Delete(key)
	return c.Get(key)
}

// Contains 判断 key 是否存在且未过期（错误标记项也算存在，只要还没过期）
func (c *ExpiredCache[K, V]) Contains(key K) bool {
	now := time.Now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.data[key]
	return ok && now.Before(item.expireAt)
}

// Len 返回缓存项数量（含未过期的错误标记项）
func (c *ExpiredCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Clear 清空缓存
func (c *ExpiredCache[K, V]) Clear() {
	c.mu.Lock()
	c.data = make(map[K]*expiredItem[V])
	c.mu.Unlock()
}

// Keys 返回所有未过期的 key
func (c *ExpiredCache[K, V]) Keys() []K {
	now := time.Now()
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, len(c.data))
	for k, item := range c.data {
		if now.Before(item.expireAt) {
			keys = append(keys, k)
		}
	}
	return keys
}

// PurgeExpired 清理所有已过期的缓存项（值和错误标记都清）
func (c *ExpiredCache[K, V]) PurgeExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, item := range c.data {
		if now.After(item.expireAt) {
			delete(c.data, k)
		}
	}
}

// SetTTL 设置缓存失效时间（只影响新写入/刷新的条目，已在缓存中的条目保持原 expireAt）
func (c *ExpiredCache[K, V]) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ttl = ttl
}

// SetLoader 设置加载回调函数
func (c *ExpiredCache[K, V]) SetLoader(loader LoaderFunc[K, V]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loader = loader
}
