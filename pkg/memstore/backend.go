package memstore

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// ============ 内存实现 ============

type stringEntry struct {
	value    string
	expireAt time.Time // 零值表示永不过期
}

// memoryStore 基于 sync.Mutex 的线程安全内存缓存
// - map + TTL 后台清理协程实现过期
// - 支持 string / hash 两种数据结构
type memoryStore struct {
	mu      sync.Mutex
	strings map[string]stringEntry
	hashes  map[string]map[string]string
}

func newMemoryStore() *memoryStore {
	m := &memoryStore{
		strings: make(map[string]stringEntry),
		hashes:  make(map[string]map[string]string),
	}
	go m.cleanupLoop()
	return m
}

// cleanupLoop 后台每 30s 扫一次过期键
func (m *memoryStore) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for k, v := range m.strings {
			if !v.expireAt.IsZero() && v.expireAt.Before(now) {
				delete(m.strings, k)
			}
		}
		m.mu.Unlock()
	}
}

func (m *memoryStore) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strings = nil
	m.hashes = nil
}

func (m *memoryStore) Ping() error { return nil }

// --- String ---

func (m *memoryStore) isExpiredRO(key string) bool {
	e, ok := m.strings[key]
	if !ok {
		return false
	}
	if e.expireAt.IsZero() {
		return false
	}
	return e.expireAt.Before(time.Now())
}

func (m *memoryStore) purgeExpired(key string) {
	if m.isExpiredRO(key) {
		delete(m.strings, key)
	}
}

func (m *memoryStore) Set(key, value string, expire time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := stringEntry{value: value}
	if expire > 0 {
		entry.expireAt = time.Now().Add(expire)
	}
	m.strings[key] = entry
	return nil
}

func (m *memoryStore) Get(key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isExpiredRO(key) {
		delete(m.strings, key)
		return "", fmt.Errorf("key not found: %s", key)
	}
	if e, ok := m.strings[key]; ok {
		return e.value, nil
	}
	return "", fmt.Errorf("key not found: %s", key)
}

func (m *memoryStore) Del(keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.strings, k)
		delete(m.hashes, k)
	}
	return nil
}

func (m *memoryStore) Exists(keys ...string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	for _, k := range keys {
		m.purgeExpired(k)
		if _, ok := m.strings[k]; ok {
			n++
			continue
		}
		if _, ok := m.hashes[k]; ok {
			n++
		}
	}
	return n, nil
}

func (m *memoryStore) Expire(key string, expire time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.strings[key]; ok {
		if expire > 0 {
			e.expireAt = time.Now().Add(expire)
		} else {
			e.expireAt = time.Time{}
		}
		m.strings[key] = e
	}
	return nil
}

func (m *memoryStore) TTL(key string) (time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.isExpiredRO(key) {
		delete(m.strings, key)
		return -2, nil
	}
	if e, ok := m.strings[key]; ok {
		if e.expireAt.IsZero() {
			return -1, nil // 永不过期
		}
		remain := time.Until(e.expireAt)
		if remain <= 0 {
			delete(m.strings, key)
			return -2, nil
		}
		return remain, nil
	}
	return -2, nil
}

func (m *memoryStore) SetNX(key, value string, expire time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.strings[key]; ok {
		return false, nil
	}
	e := stringEntry{value: value}
	if expire > 0 {
		e.expireAt = time.Now().Add(expire)
	}
	m.strings[key] = e
	return true, nil
}

func (m *memoryStore) Incr(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	if e, ok := m.strings[key]; ok {
		fmt.Sscanf(e.value, "%d", &n)
	}
	n++
	m.strings[key] = stringEntry{value: fmt.Sprintf("%d", n)}
	return n, nil
}

func (m *memoryStore) Decr(key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var n int64
	if e, ok := m.strings[key]; ok {
		fmt.Sscanf(e.value, "%d", &n)
	}
	n--
	m.strings[key] = stringEntry{value: fmt.Sprintf("%d", n)}
	return n, nil
}

// --- Hash ---

func (m *memoryStore) HashSet(key string, values ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hashes[key]
	if !ok {
		h = make(map[string]string)
		m.hashes[key] = h
	}
	for i := 0; i+1 < len(values); i += 2 {
		field := fmt.Sprint(values[i])
		val := fmt.Sprint(values[i+1])
		h[field] = val
	}
	return nil
}

func (m *memoryStore) HashGet(key, field string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if h, ok := m.hashes[key]; ok {
		if v, ok := h[field]; ok {
			return v, nil
		}
	}
	return "", fmt.Errorf("key not found: %s > %s", key, field)
}

func (m *memoryStore) HashGetAll(key string) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.hashes[key]
	if !ok {
		return nil, nil
	}
	cp := make(map[string]string, len(h))
	for k, v := range h {
		cp[k] = v
	}
	return cp, nil
}

func (m *memoryStore) HashDel(key string, fields ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if h, ok := m.hashes[key]; ok {
		for _, f := range fields {
			delete(h, f)
		}
	}
	return nil
}

func (m *memoryStore) BatchHashGet(keys []string) ([]map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]map[string]string, len(keys))
	for i, k := range keys {
		if h, ok := m.hashes[k]; ok {
			cp := make(map[string]string, len(h))
			for f, v := range h {
				cp[f] = v
			}
			out[i] = cp
		}
	}
	return out, nil
}

// sortedKeys 调试辅助：按字母序返回所有 key
func (m *memoryStore) sortedKeys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	set := map[string]struct{}{}
	for k := range m.strings {
		set[k] = struct{}{}
	}
	for k := range m.hashes {
		set[k] = struct{}{}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ============ Redis 实现 ============

type redisStore struct {
	cli *goredis.Client
	ctx context.Context
}

func newRedisStore(addr, password string, db int) *redisStore {
	return &redisStore{
		cli: goredis.NewClient(&goredis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ctx: context.Background(),
	}
}

func (r *redisStore) Close() {
	r.cli.Close()
}

func (r *redisStore) Ping() error {
	_, err := r.cli.Ping(r.ctx).Result()
	return err
}

// --- String ---

func (r *redisStore) Set(key, value string, expire time.Duration) error {
	return r.cli.Set(r.ctx, key, value, expire).Err()
}

func (r *redisStore) Get(key string) (string, error) {
	v, err := r.cli.Get(r.ctx, key).Result()
	if err == goredis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return v, err
}

func (r *redisStore) Del(keys ...string) error {
	return r.cli.Del(r.ctx, keys...).Err()
}

func (r *redisStore) Exists(keys ...string) (int64, error) {
	return r.cli.Exists(r.ctx, keys...).Result()
}

func (r *redisStore) Expire(key string, expire time.Duration) error {
	return r.cli.Expire(r.ctx, key, expire).Err()
}

func (r *redisStore) TTL(key string) (time.Duration, error) {
	d, err := r.cli.TTL(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if d == -1 || d == -2 {
		return time.Duration(int64(d)), nil
	}
	return d, nil
}

func (r *redisStore) SetNX(key, value string, expire time.Duration) (bool, error) {
	return r.cli.SetNX(r.ctx, key, value, expire).Result()
}

func (r *redisStore) Incr(key string) (int64, error) {
	return r.cli.Incr(r.ctx, key).Result()
}

func (r *redisStore) Decr(key string) (int64, error) {
	return r.cli.Decr(r.ctx, key).Result()
}

// --- Hash ---

func (r *redisStore) HashSet(key string, values ...interface{}) error {
	return r.cli.HSet(r.ctx, key, values...).Err()
}

func (r *redisStore) HashGet(key, field string) (string, error) {
	v, err := r.cli.HGet(r.ctx, key, field).Result()
	if err == goredis.Nil {
		return "", fmt.Errorf("key not found: %s > %s", key, field)
	}
	return v, err
}

func (r *redisStore) HashGetAll(key string) (map[string]string, error) {
	m, err := r.cli.HGetAll(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}
	return m, nil
}

func (r *redisStore) HashDel(key string, fields ...string) error {
	return r.cli.HDel(r.ctx, key, fields...).Err()
}

func (r *redisStore) BatchHashGet(keys []string) ([]map[string]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	pipe := r.cli.Pipeline()
	cmds := make([]*goredis.MapStringStringCmd, len(keys))
	for i, k := range keys {
		cmds[i] = pipe.HGetAll(r.ctx, k)
	}
	if _, err := pipe.Exec(r.ctx); err != nil {
		return nil, err
	}
	out := make([]map[string]string, len(keys))
	for i, c := range cmds {
		m, err := c.Result()
		if err != nil {
			if err == goredis.Nil {
				out[i] = nil
				continue
			}
			return nil, err
		}
		if len(m) == 0 {
			out[i] = nil
		} else {
			out[i] = m
		}
	}
	return out, nil
}
