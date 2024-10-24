package accumulate

import (
	"context"
	"github.com/PaesslerAG/gval"
	"github.com/god-jason/iot-master/pkg/calc"
	"strings"
)

type Result struct {
	Target   string
	Filter   map[string]any
	Document map[string]any
}

type Field struct {
	Key   string
	Value any

	_key   gval.Evaluable
	_value gval.Evaluable
}

type Accumulation struct {
	Target   string         `json:"target"`
	Filter   map[string]any `json:"filter"`
	Document map[string]any `json:"document"`

	_target gval.Evaluable
	_filter map[string]gval.Evaluable
	_fields []*Field
}

func (a *Accumulation) Init() (err error) {
	if expr, has := strings.CutPrefix(a.Target, "="); has {
		a._target, err = calc.New(expr)
		if err != nil {
			return err
		}
	}

	a._filter = make(map[string]gval.Evaluable)
	for key, value := range a.Filter {
		if val, ok := value.(string); ok {
			if expr, has := strings.CutPrefix(val, "="); has {
				a._filter[key], err = calc.New(expr)
				if err != nil {
					return err
				}
			}
		}
	}

	for key, value := range a.Document {

		f := &Field{Key: key, Value: value}

		//键
		if expr, has := strings.CutPrefix(key, "="); has {
			f._key, err = calc.New(expr)
			if err != nil {
				return err
			}
		}

		//值
		if val, ok := value.(string); ok {
			if expr, has := strings.CutPrefix(val, "="); has {
				f._value, err = calc.New(expr)
				if err != nil {
					return err
				}
			}
		}

		a._fields = append(a._fields, f)
	}

	return nil
}

func (a *Accumulation) Evaluate(args any) (result *Result, err error) {
	var ret Result

	//目标
	if a._target != nil {
		ret.Target, err = a._target.EvalString(context.Background(), args)
		if err != nil {
			return
		}
	} else {
		ret.Target = a.Target
	}

	//过滤器
	ret.Filter = make(map[string]any)
	for key, value := range a._filter {
		if value != nil {
			ret.Filter[key], err = a._target(context.Background(), args)
			if err != nil {
				return
			}
		} else {
			ret.Filter[key] = a.Filter[key]
		}
	}

	ret.Document = make(map[string]any)

	for _, f := range a._fields {
		key := f.Key
		if f._key != nil {
			key, err = f._key.EvalString(context.Background(), args)
			if err != nil {
				return
			}
		}

		val := f.Value
		if f._value != nil {
			val, err = f._value.EvalFloat64(context.Background(), args)
			if err != nil {
				return
			}

			//路过0值
			if val.(float64) == 0.0 {
				continue
			}
		}

		ret.Document[key] = val
	}

	return &ret, nil
}
