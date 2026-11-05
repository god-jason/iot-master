package ruleengine

import (
	"fmt"
	"reflect"
	"time"

	"github.com/god-jason/iot-master/internal/types"

	"github.com/expr-lang/expr"
)

// EvaluateRule 用一份模拟数据对规则条件做一次性求值（供管理端"试跑"接口使用，不产生动作）。
// 返回条件是否命中。
func EvaluateRule(rule types.Rule, values map[string]interface{}) (bool, error) {
	if values == nil {
		values = map[string]interface{}{}
	}
	if rule.Expr != "" {
		prog, err := expr.Compile(rule.Expr, expr.Env(map[string]interface{}{}), expr.AllowUndefinedVariables())
		if err != nil {
			return false, err
		}
		env := map[string]interface{}{
			"value":     values,
			"prev":      map[string]interface{}{},
			"device_id": "test",
			"ts":        time.Now().Unix(),
		}
		out, err := expr.Run(prog, env)
		if err != nil {
			return false, err
		}
		b, _ := out.(bool)
		return b, nil
	}
	return evalConditions(rule.Conditions, rule.Logic, values, map[string]interface{}{}), nil
}

// evalConditions 结构化条件组求值。
// 每个条件先对"当前值"比较；Edge=true 时还要求"上一次值不满足"，即仅在跨越瞬间命中。
// 多个条件按 Logic(and/or) 组合，默认 and；and 遇 false、or 遇 true 均提前短路返回。
func evalConditions(conds []types.RuleCondition, logic string, value, prev map[string]interface{}) bool {
	if len(conds) == 0 {
		return false
	}
	if logic == "" {
		logic = types.RuleLogicAnd
	}

	for _, cond := range conds {
		cur := compareCondition(cond, value[cond.Field])
		if cond.Edge {
			// 边沿触发：当前满足且上一次不满足
			cur = cur && !compareCondition(cond, prev[cond.Field])
		}
		if logic == types.RuleLogicOr {
			if cur {
				return true
			}
		} else {
			if !cur {
				return false
			}
		}
	}
	// and 逻辑能走到这里说明全部条件满足；or 逻辑则说明一条都没命中
	return logic == types.RuleLogicAnd
}

// compareCondition 按操作符比较单个条件。
// 支持 between/out_of（区间）、> >= < <=（数值比较）、== !=（宽松相等）；
// 值缺失或无法解析为目标类型时一律返回 false（不命中）。
func compareCondition(cond types.RuleCondition, raw interface{}) bool {
	if raw == nil {
		return false
	}
	switch cond.Op {
	case "between", "out_of":
		// 区间类操作符：当前值必须可转数值，目标值为 [min,max] 数组
		cur, ok := toFloat(raw)
		return compareRange(cond.Op, cur, ok, cond.Value)
	case ">", ">=", "<", "<=":
		// 数值比较：两侧都必须能解析为数值，否则视为不命中
		cur, ok1 := toFloat(raw)
		target, ok2 := toFloat(cond.Value)
		if !ok1 || !ok2 {
			return false
		}
		switch cond.Op {
		case ">":
			return cur > target
		case ">=":
			return cur >= target
		case "<":
			return cur < target
		case "<=":
			return cur <= target
		}
	case "==", "!=":
		// 相等类操作符走宽松比较（数值按数值、其余按字符串）
		eq := looseEqual(raw, cond.Value)
		if cond.Op == "==" {
			return eq
		}
		return !eq
	}
	return false
}

// compareRange 区间比较（between 闭区间内 / out_of 区间外）
func compareRange(op string, cur float64, curOK bool, target interface{}) bool {
	if !curOK {
		return false
	}
	lo, hi, ok := toRange(target)
	if !ok {
		return false
	}
	// 闭区间判断：between 取区间内，out_of 取区间外
	in := cur >= lo && cur <= hi
	if op == "between" {
		return in
	}
	return !in
}

// toRange 从 []interface{} / []float64 等解析 [min,max]
func toRange(v interface{}) (float64, float64, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return 0, 0, false
	}
	if rv.Len() < 2 {
		return 0, 0, false
	}
	lo, ok1 := toFloat(rv.Index(0).Interface())
	hi, ok2 := toFloat(rv.Index(1).Interface())
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	return lo, hi, true
}

// toFloat 将 JSON 数值（float64）及常见整型/布尔转换为 float64
func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case bool:
		if n {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// looseEqual 宽松相等：数值按数值比较，其余按原生相等比较
func looseEqual(a, b interface{}) bool {
	if af, ok1 := toFloat(a); ok1 {
		if bf, ok2 := toFloat(b); ok2 {
			return af == bf
		}
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}
