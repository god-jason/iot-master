package product

import (
	"encoding/json"
	"errors"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"sync"
)

// var configs = map[string]map[string]map[string]any{} 太复杂了
var configs = map[string]any{}

var configsLock sync.RWMutex

func LoadConfig[T any](product, config string) (*T, error) {
	fn := filepath.Join(viper.GetString("data"), "product", product, config+".json")

	configsLock.RLock()
	//优先从缓存中找
	if cfg, ok := configs[fn]; ok {
		configsLock.RUnlock()
		if ret, ok := cfg.(*T); ok {
			return ret, nil
		} else {
			return nil, errors.New("配置类型不对")
		}
	}

	//切换为写锁
	configsLock.RUnlock()
	configsLock.Lock()
	defer configsLock.Unlock()

	buf, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	var data T
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}

	//缓存下来
	configs[fn] = &data

	return &data, nil
}
