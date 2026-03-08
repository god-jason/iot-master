package product

import (
	"fmt"

	"github.com/spf13/cast"
)

type Compare struct {
	Type  string  `json:"type"` //= != > >= < <= eq ne gt lt gte lte
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func (c *Compare) Evaluate(ctx map[string]any) (bool, error) {
	val, ok := ctx[c.Name]
	if !ok {
		return false, fmt.Errorf("compare evalute field %s not found", c.Name)
	}
	v, err := cast.ToFloat64E(val)
	if err != nil {
		return false, err
	}
	switch c.Type {
	case "=", "==", "eq":
		return v == c.Value, nil
	case "!=", "~=", "<>", "ne":
		return v != c.Value, nil
	case ">", "gt":
		return v > c.Value, nil
	case "<", "lt":
		return v < c.Value, nil
	case ">=", "gte":
		return v >= c.Value, nil
	case "<=", "lte":
		return v <= c.Value, nil
	default:
		return false, fmt.Errorf("unsupported compare type: %s", c.Type)
	}
}

type Validator struct {
	Type       string  `json:"type"` //compare对比， expression表达式
	Compare    Compare `json:"compare,omitempty"`
	Expression string  `json:"expression,omitempty"`
	Title      string  `json:"title,omitempty"`
	Message    string  `json:"message,omitempty"`
	Level      int     `json:"level,omitempty"`
	Delay      int64   `json:"delay,omitempty"`
	Reset      int64   `json:"reset,omitempty"`
	ResetTimes int     `json:"reset_times,omitempty"`
	Disabled   bool    `json:"disabled,omitempty"`
}
