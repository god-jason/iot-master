// Package memstore 缓存抽象层：同一套 API 支持 Redis 和内存两种后端。
//
// 当初始化时传入有效的 addr（如 "127.0.0.1:6379"）且连接成功，自动走 Redis；
// 否则或连接失败时，自动降级为纯内存实现（map + Mutex + TTL 后台清理）。
// 对外函数名采用通用缓存语义，不模仿 Redis 命令风格。
package memstore

import (
	"fmt"
	"strings"
	"time"

	"github.com/god-jason/iot-master/pkg/logger"

	goredis "github.com/redis/go-redis/v9"
)

// Store 缓存存储接口 —— 内存和 Redis 都实现这个接口
type Store interface {
	Close()
	Ping() error

	// String 操作
	Set(key, value string, expire time.Duration) error
	Get(key string) (string, error)
	Del(keys ...string) error
	Exists(keys ...string) (int64, error)
	Expire(key string, expire time.Duration) error
	SetNX(key, value string, expire time.Duration) (bool, error)
	TTL(key string) (time.Duration, error)
	Incr(key string) (int64, error)
	Decr(key string) (int64, error)

	// Hash 操作
	HashSet(key string, values ...interface{}) error
	HashGet(key, field string) (string, error)
	HashGetAll(key string) (map[string]string, error)
	HashDel(key string, fields ...string) error

	// BatchHashGet 批量获取多个 hash 的全部字段
	// 返回的切片与 keys 一一对应，某个 key 不存在时对应元素为 nil
	BatchHashGet(keys []string) ([]map[string]string, error)
}

var store Store

// ============ 初始化 ============

// Init 初始化缓存后端：优先连 Redis，失败则降级内存。
// addr 为空字符串时直接用内存模式。
func Init(addr, password string, db int) error {
	if addr == "" {
		logger.Warn("cache addr empty, using memory store")
		store = newMemoryStore()
		return nil
	}

	rs := newRedisStore(addr, password, db)
	if err := rs.Ping(); err != nil {
		logger.Warn("redis connect failed, fallback to memory", "error", err)
		store = newMemoryStore()
		return nil
	}

	logger.Info("redis connected")
	store = rs
	return nil
}

// Close 关闭缓存连接
func Close() {
	if store != nil {
		store.Close()
	}
}

// IsMemory 当前是否走内存后端（运维诊断用）
func IsMemory() bool {
	_, ok := store.(*memoryStore)
	return ok
}

// ============ String 操作 ============

func Set(key, value string, expire time.Duration) error {
	return store.Set(key, value, expire)
}

func Get(key string) (string, error) {
	return store.Get(key)
}

func Del(keys ...string) error {
	return store.Del(keys...)
}

func Exists(keys ...string) (bool, error) {
	n, err := store.Exists(keys...)
	return n > 0, err
}

func Expire(key string, expire time.Duration) error {
	return store.Expire(key, expire)
}

func SetNX(key, value string, expire time.Duration) (bool, error) {
	return store.SetNX(key, value, expire)
}

func TTL(key string) (time.Duration, error) {
	return store.TTL(key)
}

func Incr(key string) (int64, error) {
	return store.Incr(key)
}

func Decr(key string) (int64, error) {
	return store.Decr(key)
}

// Pattern 按格式生成 key，如 Pattern("device", id) → "node:device:id"
func Pattern(parts ...string) string {
	return fmt.Sprintf("node:%s", strings.Join(parts, ":"))
}

// ============ Hash 操作 ============

func HashSet(key string, values ...interface{}) error {
	return store.HashSet(key, values...)
}

func HashGet(key, field string) (string, error) {
	return store.HashGet(key, field)
}

func HashGetAll(key string) (map[string]string, error) {
	return store.HashGetAll(key)
}

func HashDel(key string, fields ...string) error {
	return store.HashDel(key, fields...)
}

// BatchHashGet 批量获取多个 hash 的全部字段
func BatchHashGet(keys []string) ([]map[string]string, error) {
	return store.BatchHashGet(keys)
}

// ============ 内部辅助 ============

// isKeyNotFound 判断是否为"键不存在"错误
func isKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	if err == goredis.Nil {
		return true
	}
	return strings.Contains(err.Error(), "key not found")
}
