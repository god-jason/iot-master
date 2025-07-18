package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/god-jason/iot-master/product"
	"strings"
	"xorm.io/xorm/schemas"
)

var modelCache = lib.CacheLoader[product.Model]{
	Timeout: 600,
	Loader: func(key string) (*product.Model, error) {
		var pm product.Model
		has, err := db.Engine().ID(key).Get(&pm)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, fmt.Errorf("empty product model %s", key)
		}

		return &pm, nil
	},
}

func InvalidModel(id string) {
	modelCache.Invalid(id)
}

func LoadModel(id string) (*product.Model, error) {
	return modelCache.Load(id)
}

var configCache = lib.CacheLoader[product.ProductConfig]{
	Timeout: 600,
	Loader: func(key string) (*product.ProductConfig, error) {
		var cfg product.ProductConfig
		ss := strings.Split(key, "/")
		has, err := db.Engine().ID(schemas.PK{ss[0], ss[1]}).Get(&cfg)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, fmt.Errorf("empty product config %s", key)
		}
		return &cfg, nil
	},
}

func InvalidConfigure(id, config string) {
	idd := id + "/" + config
	configCache.Invalid(idd)
}

func LoadConfigure(id, config string) (map[string]any, error) {
	idd := id + "/" + config

	c, err := configCache.Load(idd)
	if err != nil {
		return nil, err
	}

	return c.Content, nil
}

//
//func LoadConfig[T any](id, config string) (*T, error) {
//	idd := id + "/" + config
//
//	c, err := configCache.Load(idd)
//	if err != nil {
//		return nil, err
//	}
//
//	//这里转来转去
//	buf, err := json.Marshal(c.Content)
//	if err != nil {
//		return nil, err
//	}
//
//	var t T
//	err = json.Unmarshal(buf, &t)
//	if err != nil {
//		return nil, err
//	}
//
//	return &t, nil
//}
