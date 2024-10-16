package base

import (
	"context"
	"github.com/PaesslerAG/gval"
	"github.com/god-jason/bucket/pkg/calc"
	"github.com/god-jason/bucket/types"
	"strings"
	"time"
)

type Executor struct {
	Point string        `json:"point,omitempty"`
	Value any           `json:"value,omitempty"`
	Delay time.Duration `json:"delay,omitempty"` //延迟 ms

	_value gval.Evaluable
}

type Action struct {
	Name       string              `json:"name"`
	Label      string              `json:"label"`
	Parameters []*types.SmartField `json:"parameters,omitempty"`
	Return     []*types.SmartField `json:"return,omitempty"`
	Executors  []*Executor         `json:"executors,omitempty"`
}

func (a *Action) Init() (err error) {
	for _, e := range a.Executors {
		if str, ok := e.Value.(string); ok {
			if expr, has := strings.CutPrefix(str, "="); has {
				e._value, err = calc.New(expr)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (a *Action) GetExecutors(args any) (es []*Executor, err error) {
	for _, e := range a.Executors {
		ee := *e
		if e._value != nil {
			e.Value, err = e._value(context.Background(), args)
			if err != nil {
				return nil, err
			}
		}
		es = append(es, &ee)
	}
	return es, nil
}
