package iot

import (
	"sync"
	"time"
)

type Values struct {
	values map[string]any
	lock   sync.RWMutex
	//Updated time.Time `json:"updated"`
}

func (v *Values) Put(values map[string]any) {
	if len(values) == 0 {
		return
	}

	v.lock.Lock()
	defer v.lock.Unlock()

	if v.values == nil {
		v.values = make(map[string]any)
	}

	//逐一复制
	for key, value := range values {
		v.values[key] = value
	}

	//更新时间
	v.values["_update"] = time.Now()
}

func (v *Values) Get() map[string]any {
	return v.values
}
