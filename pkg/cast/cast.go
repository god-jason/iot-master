// Package cast 提供通用的类型转换工具，零外部依赖（仅用标准库 strconv）。
//
// 设计约定：
//   - 每个转换函数失败时返回「目标类型的零值 + error」；
//   - Must 版本在失败时直接返回调用方提供的 fallback，方便直接内联使用；
//   - 输入类型优先用 type switch 匹配；无法匹配的 interface{} 用 fmt.Sprintf 转字符串再解析；
//   - []byte 先转 string 再走同路径。
package cast

import (
	"fmt"
	"strconv"
	"strings"
)

// ToInt 将任意类型安全转换为 int。
// 支持 string / 各整数类型 / 各浮点类型 / bool / []byte。
// 非法输入返回 0 + error。
func ToInt(v any) (int, error) {
	n64, err := ToInt64(v)
	if err != nil {
		return 0, err
	}
	return int(n64), nil
}

// ToInt64 将任意类型安全转换为 int64。
func ToInt64(v any) (int64, error) {
	switch x := v.(type) {
	case nil:
		return 0, fmt.Errorf("cannot convert nil to int64")
	case int:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case int64:
		return x, nil
	case uint:
		return int64(x), nil
	case uint8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case uint64:
		return int64(x), nil
	case float32:
		return int64(x), nil
	case float64:
		return int64(x), nil
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	case []byte:
		return parseIntString(string(x))
	case string:
		return parseIntString(x)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// parseIntString 先去空白，再尝试十进制整数解析。
func parseIntString(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string cannot be converted to int64")
	}
	// 兼容浮点数形式的整数字符串（如 "12.0"）
	if strings.ContainsAny(s, ".eE") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("parse float for int64: %w", err)
		}
		return int64(f), nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse int64 %q: %w", s, err)
	}
	return n, nil
}

// ToUint64 将任意类型安全转换为 uint64（负数返回错误，避免回绕）。
func ToUint64(v any) (uint64, error) {
	switch x := v.(type) {
	case nil:
		return 0, fmt.Errorf("cannot convert nil to uint64")
	case int, int8, int16, int32, int64:
		n, _ := ToInt64(x)
		if n < 0 {
			return 0, fmt.Errorf("cannot convert negative %d to uint64", n)
		}
		return uint64(n), nil
	case uint:
		return uint64(x), nil
	case uint8:
		return uint64(x), nil
	case uint16:
		return uint64(x), nil
	case uint32:
		return uint64(x), nil
	case uint64:
		return x, nil
	case float32:
		if x < 0 {
			return 0, fmt.Errorf("cannot convert negative %v to uint64", x)
		}
		return uint64(x), nil
	case float64:
		if x < 0 {
			return 0, fmt.Errorf("cannot convert negative %v to uint64", x)
		}
		return uint64(x), nil
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	case []byte:
		return parseUintString(string(x))
	case string:
		return parseUintString(x)
	default:
		return 0, fmt.Errorf("cannot convert %T to uint64", v)
	}
}

func parseUintString(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string cannot be converted to uint64")
	}
	if strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("negative string %q cannot be converted to uint64", s)
	}
	if strings.ContainsAny(s, ".eE") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("parse float for uint64: %w", err)
		}
		if f < 0 {
			return 0, fmt.Errorf("negative %v for uint64", f)
		}
		return uint64(f), nil
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse uint64 %q: %w", s, err)
	}
	return n, nil
}

// ToUint 将任意类型安全转换为 uint。
func ToUint(v any) (uint, error) {
	n, err := ToUint64(v)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}

// ToFloat64 将任意类型安全转换为 float64。
func ToFloat64(v any) (float64, error) {
	switch x := v.(type) {
	case nil:
		return 0, fmt.Errorf("cannot convert nil to float64")
	case int:
		return float64(x), nil
	case int8:
		return float64(x), nil
	case int16:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint8:
		return float64(x), nil
	case uint16:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	case bool:
		if x {
			return 1, nil
		}
		return 0, nil
	case []byte:
		return parseFloatString(string(x))
	case string:
		return parseFloatString(x)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// parseFloatString 先去空白再解析。
func parseFloatString(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string cannot be converted to float64")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float64 %q: %w", s, err)
	}
	return f, nil
}

// ToBool 将任意类型安全转换为 bool。
// 字符串支持 strconv.ParseBool 兼容集合 (1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False)
// 及扩展 yes/no、on/off、y/n（大小写不敏感）。
func ToBool(v any) (bool, error) {
	switch x := v.(type) {
	case nil:
		return false, fmt.Errorf("cannot convert nil to bool")
	case bool:
		return x, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		n, _ := ToInt64(x)
		return n != 0, nil
	case float32:
		return x != 0, nil
	case float64:
		return x != 0, nil
	case []byte:
		return parseBoolString(string(x))
	case string:
		return parseBoolString(x)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// parseBoolString 解析布尔字符串，先尝试 strconv.ParseBool，再尝试扩展集合。
func parseBoolString(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return false, fmt.Errorf("empty string cannot be converted to bool")
	}
	if b, err := strconv.ParseBool(s); err == nil {
		return b, nil
	}
	switch strings.ToLower(s) {
	case "yes", "y", "on":
		return true, nil
	case "no", "n", "off":
		return false, nil
	}
	return false, fmt.Errorf("parse bool %q: invalid syntax", s)
}

// ToString 将任意类型安全转换为 string。
func ToString(v any) (string, error) {
	switch x := v.(type) {
	case nil:
		return "", fmt.Errorf("cannot convert nil to string")
	case string:
		return x, nil
	case []byte:
		return string(x), nil
	case bool:
		return strconv.FormatBool(x), nil
	case int:
		return strconv.Itoa(x), nil
	case int8:
		return strconv.FormatInt(int64(x), 10), nil
	case int16:
		return strconv.FormatInt(int64(x), 10), nil
	case int32:
		return strconv.FormatInt(int64(x), 10), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case uint:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(x), 10), nil
	case uint64:
		return strconv.FormatUint(x, 10), nil
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", v)
	}
}

// MustInt 失败时返回 fallback。
func MustInt(v any, fallback int) int {
	n, err := ToInt(v)
	if err != nil {
		return fallback
	}
	return n
}

// MustInt64 失败返回 fallback。
func MustInt64(v any, fallback int64) int64 {
	n, err := ToInt64(v)
	if err != nil {
		return fallback
	}
	return n
}

// MustUint64 失败返回 fallback。
func MustUint64(v any, fallback uint64) uint64 {
	n, err := ToUint64(v)
	if err != nil {
		return fallback
	}
	return n
}

// MustFloat64 失败返回 fallback。
func MustFloat64(v any, fallback float64) float64 {
	f, err := ToFloat64(v)
	if err != nil {
		return fallback
	}
	return f
}

// MustString 失败时返回 fallback。
func MustString(v any, fallback string) string {
	s, err := ToString(v)
	if err != nil {
		return fallback
	}
	return s
}

// MustBool 失败时返回 fallback。
func MustBool(v any, fallback bool) bool {
	b, err := ToBool(v)
	if err != nil {
		return fallback
	}
	return b
}
