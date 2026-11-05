package iot

import (
	"fmt"
	"sync"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/cache"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"

	"xorm.io/builder"
)

// modelErrTTL 加载失败缓存的短 TTL：超过该时间后允许重新回源，
// 避免模型修正后仍被永久负缓存、必须重启进程才能生效。
const modelErrTTL = 30 * time.Second

// modelErrAt 记录各模型加载失败的时间，与模型缓存的错误标记配合实现短 TTL
var (
	modelErrMu sync.Mutex
	modelErrAt = make(map[string]time.Time)
)

// markModelErr 记录一次加载失败时间
func markModelErr(key string) {
	modelErrMu.Lock()
	modelErrAt[key] = time.Now()
	modelErrMu.Unlock()
}

// clearModelErr 清除加载失败时间记录
func clearModelErr(key string) {
	modelErrMu.Lock()
	delete(modelErrAt, key)
	modelErrMu.Unlock()
}

// modelErrStale 判断失败缓存是否已超过短 TTL
func modelErrStale(key string) bool {
	modelErrMu.Lock()
	defer modelErrMu.Unlock()
	at, ok := modelErrAt[key]
	return !ok || time.Since(at) >= modelErrTTL
}

// models 物模型缓存（运行时 thing.Model）
// loader 从 DB 查 types.Model 后取内嵌的 m.Model 返回
var models = cache.NewCache[string, *thing.Model](func(key string) (*thing.Model, error) {
	// 缓存未命中时的 loader：从 DB 查 types.Model，取内嵌的运行时模型 m.Model
	var m types.Model
	found, err := db.FindOne("model", builder.Eq{"id": key}, &m)
	if err != nil {
		markModelErr(key)
		return nil, fmt.Errorf("iot/models: model %s DB 错误: %w", key, err)
	}
	if !found || m.ID == "" {
		markModelErr(key)
		return nil, fmt.Errorf("iot/models: model %s 不存在", key)
	}
	return &m.Model, nil
})

// Models 返回缓存（方便批量操作）
func Models() *cache.Cache[string, *thing.Model] {
	return models
}

// GetModel 获取物模型（运行时 thing.Model）
//   - cache hit 且无 err → 直接返回
//   - cache hit 但有 err → 超过短 TTL（modelErrTTL）后清除错误缓存并重新加载，否则返回 nil,false
//   - cache miss → loader 从 DB 加载（singleflight）
func GetModel(modelId string) (*thing.Model, bool) {
	m, err := models.Get(modelId)
	if err == nil && m != nil {
		clearModelErr(modelId)
		return m, true
	}
	// 命中错误缓存：未超过短 TTL 直接视为未找到；超过则清缓存重新回源（singleflight 由缓存内部保证）
	if err != nil && modelErrStale(modelId) {
		models.Delete(modelId)
		clearModelErr(modelId)
		m, err = models.Get(modelId)
		if err == nil && m != nil {
			return m, true
		}
	}
	return nil, false
}

// SetModel 注册/覆盖物模型（同时缓存 thing.Model + 触发协议归一化）
// 清除可能存在的 err 标记
func SetModel(modelId string, m *thing.Model) {
	models.Set(modelId, m)
	clearModelErr(modelId)
	_ = protocol.RegisterModel(modelId, m)
}

// DeleteModel 从缓存移除
func DeleteModel(modelId string) {
	models.Delete(modelId)
	clearModelErr(modelId)
	protocol.UnregisterModel(modelId)
}

// ClearModels 清空所有缓存
func ClearModels() {
	models.Clear()
	modelErrMu.Lock()
	modelErrAt = make(map[string]time.Time)
	modelErrMu.Unlock()
}
