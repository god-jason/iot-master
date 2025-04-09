package product

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"strings"
	"xorm.io/xorm/schemas"
)

var configCache = lib.CacheLoader[ProductConfig]{
	Timeout: 600,
	Loader: func(key string) (*ProductConfig, error) {
		var cfg ProductConfig
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

func LoadConfig[T any](id, config string) (*T, error) {
	idd := id + "/" + config

	c, err := configCache.Load(idd)
	if err != nil {
		return nil, err
	}

	//这里转来转去
	buf, err := json.Marshal(c.Content)
	if err != nil {
		return nil, err
	}

	var t T
	err = json.Unmarshal(buf, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
