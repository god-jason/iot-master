package aggregator

import (
	"github.com/spf13/cast"
)

type avg struct {
	value  float64
	count  float64
	result float64
}

func (a *avg) Push(value any) error {
	res, err := cast.ToFloat64E(value)
	if err != nil {
		return err
	}
	a.value += res
	a.count++
	return nil
}

func (a *avg) Snap() {
	if a.count == 0 {
		return
	}
	a.result = a.value / a.count
	a.value = 0
	a.count = 0
	return
}

func (a *avg) Pop() any {
	return a.result
}
