package protocol

import (
	"fmt"
	"sync"

	"github.com/god-jason/iot-master/pkg/thing"
)

var models sync.Map

// RegisterModel 全局注册物模型
// modelId 是缓存 key（由调用方提供，通常是 DB Model.ID 或网关生成的唯一标识）
// 根据 model.Protocol 找到对应 Manifest.Normalizer，把 DB 原始模型归一化成协议专属模型后缓存
// 所有 Protocol 实例（跨连接）共享同一份解析后的模型
func RegisterModel(modelId string, model *thing.Model) error {
	if model == nil {
		return fmt.Errorf("protocol: model 为 nil")
	}
	manifest := GetManifest(model.Protocol)
	if manifest.Normalizer == nil {
		return fmt.Errorf("protocol: 协议 %q 未注册 Normalizer", model.Protocol)
	}
	realModel, err := manifest.Normalizer(model)
	if err != nil {
		return fmt.Errorf("protocol: 模型 %s 归一化失败: %w", modelId, err)
	}
	// 回填 modelId 到协议专属模型（如果它有 Id 字段）
	if m, ok := realModel.(interface{ SetId(string) }); ok {
		m.SetId(modelId)
	}
	models.Store(modelId, realModel)
	return nil
}

// GetModel 获取已注册的模型（解析后的协议专属类型，如 *ModbusModel / *dltCachedModel 等）
func GetModel(modelId string) (any, bool) {
	return models.Load(modelId)
}

// UnregisterModel 注销模型
func UnregisterModel(modelId string) {
	models.Delete(modelId)
}
