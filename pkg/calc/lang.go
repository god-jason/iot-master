package calc

import (
	"math"

	"github.com/PaesslerAG/gval"
	"github.com/spf13/cast"
)

func tarFunc(fun func(x float64) float64) func(x any) float64 {
	return func(x any) float64 {
		xx := cast.ToFloat64(x)
		return fun(xx)
	}
}

func tarFunc2(fun func(x, y float64) float64) func(x, y any) float64 {
	return func(x, y any) float64 {
		xx := cast.ToFloat64(x)
		yy := cast.ToFloat64(y)
		return fun(xx, yy)
	}
}

// 添加了数据函数的gval
var lang = gval.Full(
	gval.Constant("E", math.E),
	gval.Constant("LN10", math.Ln10),
	gval.Constant("LN2", math.Ln2),
	gval.Constant("LOG10E", math.Log2E),
	gval.Constant("LOG2E", math.Log10E),
	gval.Constant("PI", math.Pi),
	//gval.Constant("PI", math.Sqrt2),
	//gval.Constant("PI", math.SqrtE),
	gval.Function("ABS", tarFunc(math.Abs)),
	gval.Function("CEIL", tarFunc(math.Ceil)),
	gval.Function("FLOOR", tarFunc(math.Floor)),
	gval.Function("TRUNC", tarFunc(math.Trunc)),
	gval.Function("POW", tarFunc2(math.Pow)),
	gval.Function("ROUND", tarFunc(math.Round)),
	gval.Function("SQRT", tarFunc(math.Sqrt)),
	gval.Function("CBRT", tarFunc(math.Cbrt)),
	gval.Function("EXP", tarFunc(math.Exp)),
	gval.Function("EXP2", tarFunc(math.Exp2)),
	gval.Function("EXPM1", tarFunc(math.Expm1)),
	gval.Function("SIN", tarFunc(math.Sin)),
	gval.Function("SINH", tarFunc(math.Sinh)),
	gval.Function("ASIN", tarFunc(math.Asin)),
	gval.Function("ASINH", tarFunc(math.Asinh)),
	gval.Function("COS", tarFunc(math.Cos)),
	gval.Function("COSH", tarFunc(math.Cosh)),
	gval.Function("ACOS", tarFunc(math.Acos)),
	gval.Function("ACOSH", tarFunc(math.Acosh)),
	gval.Function("TAN", tarFunc(math.Tan)),
	gval.Function("TANH", tarFunc(math.Tanh)),
	gval.Function("ATAN", tarFunc(math.Atan)),
	gval.Function("ATANH", tarFunc(math.Atanh)),
	gval.Function("LOG", tarFunc(math.Log)),
	gval.Function("LOG2", tarFunc(math.Log2)),
	gval.Function("LOG10", tarFunc(math.Log10)),
	gval.Function("LOG1p", tarFunc(math.Log1p)),
	gval.Function("HYPOT", tarFunc2(math.Hypot)),
	gval.Function("MAX", tarFunc2(math.Max)),
	gval.Function("MIN", tarFunc2(math.Min)),
	//gval.Function("RANDOM", func() {}),

)

func Compile(expression string) (gval.Evaluable, error) {
	return lang.NewEvaluable(expression)
}
