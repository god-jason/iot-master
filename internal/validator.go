package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PaesslerAG/gval"
	"github.com/god-jason/iot-master/calc"
	"github.com/god-jason/iot-master/product"
	"github.com/spf13/cast"
)

type Validator struct {
	*product.Validator

	expression gval.Evaluable

	start int64
	times int
}

func (v *Validator) Build() (err error) {
	if v.Type == "expression" && v.Expression != "" {
		v.expression, err = calc.Compile(v.Expression)
	}
	return err
}

func (v *Validator) Evaluate(ctx map[string]any) (*Alarm, error) {
	var err error
	var ret bool

	//检查条件
	switch v.Type {
	case "compare":
		ret, err = v.Compare.Evaluate(ctx)
	case "expression":
		if v.expression == nil {
			return nil, fmt.Errorf("invalid compare expression")
		}
		ret, err = v.expression.EvalBool(context.Background(), ctx)
	default:
		err = fmt.Errorf("unsupported validator type: %s", v.Type)
	}

	if err != nil {
		return nil, err
	}

	//条件为真时，产生报警，所以条件为假，则重置
	if !ret {
		v.start = 0
		v.times = 0
		return nil, nil
	}

	//避免重复报警
	if v.times > 0 && v.Reset <= 0 {
		return nil, nil
	}

	//起始时间
	now := time.Now().Unix()
	if v.start == 0 {
		v.start = now
	}

	//延迟报警
	if v.Delay > 0 {
		if now < v.start+v.Delay {
			return nil, nil
		}
	}

	if v.times > 0 {
		//重复报警
		if v.Reset <= 0 {
			return nil, nil
		}

		//超过最大次数
		if v.ResetTimes > 0 && v.times > v.ResetTimes {
			return nil, nil
		}

		//还没到时间
		if now < v.start+v.Reset {
			return nil, nil
		}

		//重置开始时间
		v.start = now
	}

	v.times = v.times + 1

	//产生报警
	a := &Alarm{
		Title:   replaceParams(v.Title, ctx),
		Message: replaceParams(v.Message, ctx),
		Level:   v.Level,
	}

	return a, nil
}

var paramsRegex *regexp.Regexp

func init() {
	paramsRegex = regexp.MustCompile(`\{\w+\}`)
}

func replaceParams(str string, ctx map[string]any) string {
	return paramsRegex.ReplaceAllStringFunc(str, func(s string) string {
		s = strings.TrimPrefix(s, "{")
		s = strings.TrimSuffix(s, "}")
		return cast.ToString(ctx[s])
	})
}
