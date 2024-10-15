package condition

import (
	"context"
	"github.com/PaesslerAG/gval"
	"github.com/zgwit/iot-master/pkg/calc"
)

type Compare struct {
	Variable string `json:"variable,omitempty"` //变量
	Operator string `json:"operator,omitempty"` //对比算子 > >= < <= !=
	Value    string `json:"value,omitempty"`    //值，支持表达式
	_value   gval.Evaluable
}

func (c *Compare) Init() (err error) {
	c._value, err = calc.New(c.Variable + c.Operator + "(" + c.Value + ")")
	return
}

func (c *Compare) Eval(ctx map[string]any) (bool, error) {
	return c._value.EvalBool(context.Background(), ctx)
}
