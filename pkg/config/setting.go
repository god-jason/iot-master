package config

import (
	"sync"

	"github.com/god-jason/iot-master/pkg/smart"
)

type Form struct {
	Title  string        `json:"title"`
	Module string        `json:"module"`
	Fields []smart.Field `json:"fields"`
}

var modules sync.Map

func Register(module string, form *Form) {
	modules.Store(module, form)
}

func Unregister(module string) {
	modules.Delete(module)
}

func GetModule(module string) *Form {
	if v, ok := modules.Load(module); ok {
		return v.(*Form)
	}
	return nil
}

func GetModules() []Form {
	var ms []Form
	modules.Range(func(key, value any) bool {
		m := *value.(*Form)
		m.Fields = nil
		ms = append(ms, m)
		return true
	})
	return ms
}
