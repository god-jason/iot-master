package cast

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// ToDuration 将任意类型转换为 time.Duration。
// string 支持：纯数字（纳秒）、"1h30m" 等 Go duration 字符串、float 秒。
func ToDuration(v any) (time.Duration, error) {
	switch x := v.(type) {
	case time.Duration:
		return x, nil
	case int, int8, int16, int32, int64:
		n, _ := ToInt64(x)
		return time.Duration(n), nil
	case float32, float64:
		f, _ := ToFloat64(x)
		return time.Duration(f), nil
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, fmt.Errorf("empty string cannot be converted to duration")
		}
		// 纯数字 → 纳秒
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return time.Duration(n), nil
		}
		// 数字 + 单位 → Go duration
		if d, err := time.ParseDuration(s); err == nil {
			return d, nil
		}
		// 浮点秒
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return time.Duration(f * float64(time.Second)), nil
		}
		return 0, fmt.Errorf("parse duration %q failed", s)
	default:
		return 0, fmt.Errorf("cannot convert %T to duration", v)
	}
}

// MustDuration 失败返回 fallback。
func MustDuration(v any, fallback time.Duration) time.Duration {
	d, err := ToDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// ToTime 将任意类型转换为 time.Time。
// 支持 RFC3339 / RFC3339Nano 字符串、Unix 秒/毫秒/微秒/纳秒整数或浮点。
func ToTime(v any) (time.Time, error) {
	switch x := v.(type) {
	case time.Time:
		return x, nil
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return time.Time{}, fmt.Errorf("empty string cannot be converted to time")
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, s); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("parse time %q failed", s)
	case int, int8, int16, int32, int64:
		return unixToTime(ToInt64(x))
	case uint, uint8, uint16, uint32, uint64:
		n, _ := ToUint64(x)
		return time.Unix(int64(n), 0), nil
	case float32, float64:
		f, _ := ToFloat64(x)
		sec := int64(f)
		nsec := int64((f - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", v)
	}
}

// unixToTime 根据数值量级推断单位（秒/毫秒/微秒/纳秒）。
func unixToTime(n int64, err error) (time.Time, error) {
	if err != nil {
		return time.Time{}, err
	}
	switch {
	case n > 1e17: // 纳秒
		return time.Unix(0, n), nil
	case n > 1e14: // 微秒
		return time.Unix(0, n*1e3), nil
	case n > 1e11: // 毫秒
		return time.Unix(0, n*1e6), nil
	default: // 秒
		return time.Unix(n, 0), nil
	}
}

// MustTime 失败返回 fallback。
func MustTime(v any, fallback time.Time) time.Time {
	t, err := ToTime(v)
	if err != nil {
		return fallback
	}
	return t
}

// ToStringSlice 将任意类型转换为 []string。
// 支持 []string / []any（逐元素 ToString），兼容逗号分隔的单字符串。
func ToStringSlice(v any) ([]string, error) {
	switch x := v.(type) {
	case nil:
		return nil, fmt.Errorf("cannot convert nil to string slice")
	case []string:
		return x, nil
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			s, err := ToString(e)
			if err != nil {
				return nil, err
			}
			out = append(out, s)
		}
		return out, nil
	case string:
		if strings.TrimSpace(x) == "" {
			return []string{}, nil
		}
		parts := strings.Split(x, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to string slice", v)
	}
}

// ToIntSlice 将任意类型转换为 []int。
func ToIntSlice(v any) ([]int, error) {
	ss, err := ToStringSlice(v)
	if err != nil {
		// 可能是数值切片
		switch x := v.(type) {
		case []int:
			return x, nil
		case []any:
			out := make([]int, 0, len(x))
			for _, e := range x {
				n, err := ToInt(e)
				if err != nil {
					return nil, err
				}
				out = append(out, n)
			}
			return out, nil
		}
		return nil, err
	}
	out := make([]int, 0, len(ss))
	for _, s := range ss {
		n, err := ToInt(s)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

// ToStringMap 将 map[string]any 外的常见 map 类型转换为 map[string]any（浅拷贝）。
func ToStringMap(v any) (map[string]any, error) {
	switch m := v.(type) {
	case nil:
		return nil, fmt.Errorf("cannot convert nil to map")
	case map[string]any:
		return m, nil
	case map[string]string:
		out := make(map[string]any, len(m))
		for k, val := range m {
			out[k] = val
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to map[string]any", v)
	}
}

// ToAddr 将 "host:port" / "host" / netip.Addr 转换为 netip.AddrPort。
// 缺端口时默认 0；仅用于内部网络工具。
func ToAddr(v any) (netip.AddrPort, error) {
	switch x := v.(type) {
	case netip.AddrPort:
		return x, nil
	case netip.Addr:
		return netip.AddrPortFrom(x, 0), nil
	case string:
		if ap, err := netip.ParseAddrPort(x); err == nil {
			return ap, nil
		}
		if a, err := netip.ParseAddr(x); err == nil {
			return netip.AddrPortFrom(a, 0), nil
		}
		return netip.AddrPort{}, fmt.Errorf("parse addr %q failed", x)
	default:
		return netip.AddrPort{}, fmt.Errorf("cannot convert %T to netip.AddrPort", v)
	}
}
