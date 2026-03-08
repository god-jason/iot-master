package config

import (
	"github.com/god-jason/iot-master/pkg/lib"
	"github.com/god-jason/iot-master/pkg/smart"
)

type Form struct {
	Title  string        `json:"title"`
	Module string        `json:"module"`
	Fields []smart.Field `json:"fields"`
}

var modules lib.Map[Form]

func Register(module string, form *Form) {
	modules.Store(module, form)
}

func Unregister(module string) {
	modules.Delete(module)
}

func GetModule(module string) *Form {
	return modules.Load(module)
}

func GetModules() []Form {
	var ms []Form
	modules.Range(func(_ string, item *Form) bool {
		m := *item
		m.Fields = nil
		ms = append(ms, m)
		return true
	})
	return ms
}
