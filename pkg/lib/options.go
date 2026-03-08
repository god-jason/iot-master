package lib

import (
	"github.com/spf13/cast"
)

type Options map[string]any

func (o Options) Float64(name string, def float64) float64 {
	if value, ok := o[name]; ok {
		return cast.ToFloat64(value)
	}
	return def
}

func (o Options) Int64(name string, def int64) int64 {
	if value, ok := o[name]; ok {
		return cast.ToInt64(value)
	}
	return def
}

func (o Options) Int(name string, def int) int {
	if value, ok := o[name]; ok {
		return cast.ToInt(value)
	}
	return def
}

func (o Options) Bool(name string, def bool) bool {
	if value, ok := o[name]; ok {
		return cast.ToBool(value)
	}
	return def
}
