package cache

import (
	"sync"
)

// cacheItem 缓存项：存值 + 上次加载错误（error != nil 表示曾加载失败，后续不再自动重试）
type cacheItem[V any] struct {
	value V
	err   error
}

// Cache 泛型常驻缓存，无 TTL、无后台清理
//
//   - 快路径：RWMutex 读锁直接命中
//   - 慢路径：singleflight 避免并发击穿，同 key 多协程共享一次 loader 调用
//   - 错误缓存：loader 返回 err 时同 key 后续 Get 直接返回 err，避免反复查 DB；
//     调用方显式 Set/Delete/Refresh 或换 Loader 后才会重新加载
//
// 典型场景：物模型、设备分组等低频更新、高频读取的进程内字典。
// 设计参考 iot/modelcache.go。
type Cache[K comparable, V any] struct {
	data   map[K]*cacheItem[V]
	loader LoaderFunc[K, V]

	mu       sync.RWMutex          // 保护 data
	flight   sync.Mutex            // 保护 inflight
	inflight map[K]*sync.WaitGroup // singleflight：同 key 并发请求共享一次 loader
}

// NewCache 创建常驻缓存，loader 可通过 SetLoader 后置注入
func NewCache[K comparable, V any](loader LoaderFunc[K, V]) *Cache[K, V] {
	return &Cache[K, V]{
		data:     make(map[K]*cacheItem[V]),
		loader:   loader,
		inflight: make(map[K]*sync.WaitGroup),
	}
}

// Get 获取缓存值
//
//	cache hit + err=nil     → 直接返回
//	cache hit + err!=nil    → 返回之前的错误（不自动重试）
//	cache miss              → singleflight 调用 loader 加载并落缓存
func (c *Cache[K, V]) Get(key K) (V, error) {
	// 1. 快路径：读锁命中
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()
	if ok {
		var zero V
		if item.err != nil {
			return zero, item.err
		}
		return item.value, nil
	}

	// 2. 慢路径：singleflight
	c.flight.Lock()
	wg, existing := c.inflight[key]
	if !existing {
		wg = &sync.WaitGroup{}
		wg.Add(1)
		c.inflight[key] = wg
	}
	c.flight.Unlock()

	if existing {
		// 有兄弟协程在加载，等它
		wg.Wait()
		c.mu.RLock()
		item, ok = c.data[key]
		c.mu.RUnlock()
		if ok {
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
	c.mu.Lock()
	c.data[key] = &cacheItem[V]{value: v, err: err}
	c.mu.Unlock()
	return v, err
}

// Set 手动设置缓存值（同时清掉可能存在的 err 标记）
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	c.data[key] = &cacheItem[V]{value: value}
	c.mu.Unlock()
}

// SetErr 标记该 key 已加载失败（后续 Get 会直接返回该 err，不再重试）
// 用于主动注入一个错误状态，避免同 key 被反复请求击穿
func (c *Cache[K, V]) SetErr(key K, err error) {
	var zero V
	c.mu.Lock()
	c.data[key] = &cacheItem[V]{value: zero, err: err}
	c.mu.Unlock()
}

// Delete 从缓存移除（包括错误标记）；下次 Get 会走 loader
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

// Refresh 强制重新加载（忽略旧值和错误标记）
func (c *Cache[K, V]) Refresh(key K) (V, error) {
	c.Delete(key)
	return c.Get(key)
}

// Contains 判断 key 是否在缓存中（包括错误标记项也算存在）
func (c *Cache[K, V]) Contains(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.data[key]
	return ok
}

// Len 返回缓存项数量
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Clear 清空全部缓存
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	c.data = make(map[K]*cacheItem[V])
	c.mu.Unlock()
}

// Keys 返回所有 key
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// SetLoader 换一个 loader（已缓存的错误项会被保留，需手动 Delete/Refresh 才会重试）
func (c *Cache[K, V]) SetLoader(loader LoaderFunc[K, V]) {
	c.mu.Lock()
	c.loader = loader
	c.mu.Unlock()
}
