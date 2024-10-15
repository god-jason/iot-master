package aggregator

import (
	"github.com/spf13/cast"
)

type sum struct {
	value float64
	dirty bool

	result float64
}

func (a *sum) Push(value any) error {
	res, err := cast.ToFloat64E(value)
	if err != nil {
		return err
	}
	a.value += res
	a.dirty = true
	return nil
}

func (a *sum) Snap() {
	if !a.dirty {
		return
	}
	a.result = a.value
	a.value = 0
	a.dirty = false
}

func (a *sum) Pop() any {
	return a.result
}
