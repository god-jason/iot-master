package aggregator

import (
	"github.com/spf13/cast"
)

type maximum struct {
	value float64
	dirty bool

	result float64
}

func (a *maximum) Push(value any) error {
	res, err := cast.ToFloat64E(value)
	if err != nil {
		return err
	}

	if !a.dirty {
		a.value = res
	} else if res > a.value {
		a.value = res
	}
	a.dirty = true
	return nil
}

func (a *maximum) Snap() {
	if !a.dirty {
		return
	}
	a.result = a.value
	a.dirty = false
	return
}

func (a *maximum) Pop() any {
	return a.result
}
