package aggregator

import (
	"github.com/spf13/cast"
)

type inc struct {
	hasLast bool

	last    float64
	current float64

	dirty bool

	result float64
}

func (a *inc) Push(value any) error {
	res, err := cast.ToFloat64E(value)
	if err != nil {
		return err
	}
	a.current = res
	a.dirty = true

	return nil
}

func (a *inc) Snap() {
	if !a.dirty {
		return
	}

	if !a.hasLast {
		a.hasLast = true
		a.last = a.current
		a.dirty = false
		return
	}

	a.result = a.current - a.last
	a.last = a.current
	a.dirty = false
}

func (a *inc) Pop() any {
	return a.result
}
