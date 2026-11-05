package thing

import (
	"math"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/spf13/cast"
)

// compileMu 保护 Expression/Filter 的懒编译字段（program/err）。
// 同一个 *Model 会被多个设备并发使用，懒编译必须串行，避免数据竞争。
var compileMu sync.Mutex

// exprMathFuncs 用 expr.Function 注册 math 函数（编译时注入一次即可，运行时不再传递）
var exprMathFuncs = []expr.Option{
	expr.Function("ABS", func(params ...any) (any, error) {
		return math.Abs(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("MIN", func(params ...any) (any, error) {
		return math.Min(cast.ToFloat64(params[0]), cast.ToFloat64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("MAX", func(params ...any) (any, error) {
		return math.Max(cast.ToFloat64(params[0]), cast.ToFloat64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("ROUND", func(params ...any) (any, error) {
		return math.Round(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("SQRT", func(params ...any) (any, error) {
		return math.Sqrt(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("POW", func(params ...any) (any, error) {
		return math.Pow(cast.ToFloat64(params[0]), cast.ToFloat64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("CEIL", func(params ...any) (any, error) {
		return math.Ceil(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("FLOOR", func(params ...any) (any, error) {
		return math.Floor(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("LOG", func(params ...any) (any, error) {
		return math.Log(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("LOG10", func(params ...any) (any, error) {
		return math.Log10(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("SIN", func(params ...any) (any, error) {
		return math.Sin(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("COS", func(params ...any) (any, error) {
		return math.Cos(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("TAN", func(params ...any) (any, error) {
		return math.Tan(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("ATAN2", func(params ...any) (any, error) {
		return math.Atan2(cast.ToFloat64(params[0]), cast.ToFloat64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("ATAN", func(params ...any) (any, error) {
		return math.Atan(cast.ToFloat64(params[0])), nil
	}, new(func(float64) float64)),
	expr.Function("DEG", func(params ...any) (any, error) {
		return cast.ToFloat64(params[0]) * 180 / math.Pi, nil
	}, new(func(float64) float64)),
	expr.Function("RAD", func(params ...any) (any, error) {
		return cast.ToFloat64(params[0]) * math.Pi / 180, nil
	}, new(func(float64) float64)),
	// 位运算
	expr.Function("BAND", func(params ...any) (any, error) {
		return float64(cast.ToInt64(params[0]) & cast.ToInt64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("BOR", func(params ...any) (any, error) {
		return float64(cast.ToInt64(params[0]) | cast.ToInt64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("BXOR", func(params ...any) (any, error) {
		return float64(cast.ToInt64(params[0]) ^ cast.ToInt64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("BSL", func(params ...any) (any, error) {
		return float64(cast.ToInt64(params[0]) << cast.ToInt64(params[1])), nil
	}, new(func(float64, float64) float64)),
	expr.Function("BSR", func(params ...any) (any, error) {
		return float64(cast.ToInt64(params[0]) >> cast.ToInt64(params[1])), nil
	}, new(func(float64, float64) float64)),
}

// CalcExpressions 计算所有表达式，结果**直接写入 values**
// Expression.program/err 首次失败后永久标记跳过，不重试
// 编译时 env 用所有 key 占位（让 expr 编译器知道标识符存在），运行时传真实 values
func CalcExpressions(expressions []Expression, values map[string]any) {
	if len(expressions) == 0 {
		return
	}

	for i := range expressions {
		exprDef := &expressions[i] // 取地址，首次编译后原地缓存

		if exprDef.Key == "" || exprDef.Expr == "" || exprDef.err != nil {
			continue
		}

		if exprDef.program == nil {
			compileMu.Lock()
			if exprDef.program == nil {
				opts := append([]expr.Option{expr.Env(values)}, exprMathFuncs...)
				exprDef.program, exprDef.err = expr.Compile(exprDef.Expr, opts...)
			}
			compileMu.Unlock()
			if exprDef.program == nil {
				continue
			}
		}

		result, err := expr.Run(exprDef.program, values)
		if err != nil {
			// 运行时错误（临时，不标记永久跳过）
			continue
		}

		values[exprDef.Key] = result
	}
}

// PostProcess 对 protocol.Poll 返回的 values 做后处理
// model 是 DB 原始物模型（含 Expressions/Filters）
// 返回 true 表示数据通过，false 表示被 Filter 拦截
func PostProcess(model *Model, values map[string]any) bool {
	if model == nil || values == nil {
		return true
	}
	CalcExpressions(model.Expressions, values)
	return ApplyFilters(model.Filters, values)
}

// ApplyFilters 运行所有过滤器
// 返回 true 表示数据通过所有 filter，false 表示被任意一个 filter 拦截
func ApplyFilters(filters []Filter, values map[string]any) bool {
	if len(filters) == 0 {
		return true
	}

	for i := range filters {
		f := &filters[i]

		if f.Name == "" || f.Expr == "" || f.err != nil {
			continue
		}

		if f.program == nil {
			compileMu.Lock()
			if f.program == nil {
				opts := append([]expr.Option{expr.Env(values)}, exprMathFuncs...)
				f.program, f.err = expr.Compile(f.Expr, opts...)
			}
			compileMu.Unlock()
			if f.program == nil {
				continue
			}
		}

		result, err := expr.Run(f.program, values)
		if err != nil {
			continue
		}

		if !cast.ToBool(result) {
			return false
		}
	}
	return true
}

// ApplyPropertyRange 按 Property 的 Min/Max 做范围过滤
// max > min 时才检查；任一属性值越界则整体返回 false（整条丢弃）
func ApplyPropertyRange(properties []Property, values map[string]any) bool {
	for _, p := range properties {
		if p.Min >= p.Max || p.Key == "" {
			continue
		}
		v, ok := values[p.Key]
		if !ok {
			continue
		}
		fv := cast.ToFloat64(v)
		if fv < p.Min || fv > p.Max {
			return false
		}
	}
	return true
}
