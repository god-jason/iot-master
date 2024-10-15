package aggregator

import (
	"github.com/spf13/cast"
)

type minimum struct {
	value float64
	dirty bool

	result float64
}

func (a *minimum) Push(value any) error {
	res, err := cast.ToFloat64E(value)
	if err != nil {
		return err
	}

	if !a.dirty {
		a.value = res
	} else if res < a.value {
		a.value = res
	}
	a.dirty = false

	return nil
}

func (a *minimum) Snap() {
	if !a.dirty {
		return
	}
	a.result = a.value
	a.dirty = false
}

func (a *minimum) Pop() any {
	return a.result
}
