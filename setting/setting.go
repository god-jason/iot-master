package setting

import (
	"github.com/god-jason/iot-master/lib"
	"github.com/god-jason/iot-master/types"
)

type Module struct {
	Name   string             `json:"name"`
	Module string             `json:"module"`
	Title  string             `json:"title,omitempty"`
	Form   []types.SmartField `json:"-"`
}

var modules lib.Map[Module]

func Register(module string, form *Module) {
	modules.Store(module, form)
}

func Unregister(module string) {
	modules.Delete(module)
}

func Load(module string) *Module {
	return modules.Load(module)
}

func Modules() []*Module {
	var ms []*Module
	modules.Range(func(_ string, item *Module) bool {
		ms = append(ms, item)
		return true
	})
	return ms
}
