package cast

import (
	"fmt"
	"net/netip"
	"testing"
	"time"
)

var errPlaceholder = fmt.Errorf("placeholder")

// -------- ToInt / ToInt64 --------

func TestToInt64_Basic(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want int64
	}{
		{"int", 42, 42},
		{"int8", int8(-8), -8},
		{"int16", int16(16), 16},
		{"int32", int32(32), 32},
		{"int64", int64(64), 64},
		{"uint", uint(10), 10},
		{"uint8", uint8(8), 8},
		{"uint16", uint16(16), 16},
		{"uint32", uint32(32), 32},
		{"uint64", uint64(64), 64},
		{"float64", 3.14, 3},
		{"float32", float32(2.7), 2},
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"string int", "  123  ", 123},
		{"string float whole", "12.0", 12},
		{"string sci", "1e2", 100},
		{"bytes", []byte("-7"), -7},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToInt64(c.in)
			if err != nil {
				t.Fatalf("ToInt64(%v) unexpected err: %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestToInt64_Error(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{"nil", nil},
		{"bad string", "abc"},
		{"empty string", ""},
		{"bad float str", "1.2.3"},
		{"struct", struct{}{}},
		{"func", func() {}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ToInt64(c.in); err == nil {
				t.Fatalf("ToInt64(%v) expected error", c.in)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	n, err := ToInt(int64(1<<31 - 1))
	if err != nil {
		t.Fatal(err)
	}
	if n != 1<<31-1 {
		t.Fatalf("got %d", n)
	}
	if _, err := ToInt("bad"); err == nil {
		t.Fatal("expected error")
	}
}

// -------- ToUint64 / ToUint --------

func TestToUint64_Basic(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want uint64
	}{
		{"uint", uint(10), 10},
		{"uint8", uint8(8), 8},
		{"uint16", uint16(16), 16},
		{"uint32", uint32(32), 32},
		{"uint64", uint64(64), 64},
		{"int positive", 5, 5},
		{"int8", int8(8), 8},
		{"float", 3.9, 3},
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"string", "123", 123},
		{"string float", "45.0", 45},
		{"string e", "1e2", 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToUint64(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %d want %d", got, c.want)
			}
		})
	}
}

func TestToUint64_Error(t *testing.T) {
	cases := []any{nil, -1, "-5", "abc", "", -0.1, struct{}{}, "-3.5", "notnum.xx"}
	for _, v := range cases {
		if _, err := ToUint64(v); err == nil {
			t.Fatalf("ToUint64(%v) expected error", v)
		}
	}
}

func TestToUint(t *testing.T) {
	n, err := ToUint("42")
	if err != nil || n != 42 {
		t.Fatalf("got %d %v", n, err)
	}
}

// -------- ToFloat64 --------

func TestToFloat64_Basic(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want float64
	}{
		{"int", 3, 3},
		{"int8", int8(-1), -1},
		{"int16", int16(16), 16},
		{"int32", int32(32), 32},
		{"uint", uint(10), 10},
		{"uint8", uint8(8), 8},
		{"uint16", uint16(16), 16},
		{"uint32", uint32(32), 32},
		{"uint64", uint64(64), 64},
		{"float32", float32(2.5), 2.5},
		{"float64", 3.14, 3.14},
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"string", "2.5", 2.5},
		{"string int", "7", 7},
		{"bytes", []byte("1.25"), 1.25},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToFloat64(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestToFloat64_Error(t *testing.T) {
	for _, v := range []any{nil, "", "abc", struct{}{}} {
		if _, err := ToFloat64(v); err == nil {
			t.Fatalf("ToFloat64(%v) expected error", v)
		}
	}
}

// -------- ToBool --------

func TestToBool_Basic(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int non-zero", 1, true},
		{"int zero", 0, false},
		{"int neg", -5, true},
		{"uint zero", uint(0), false},
		{"float non-zero", 0.1, true},
		{"float zero", 0.0, false},
		{"str 1", "1", true},
		{"str T", "T", true},
		{"str true", "true", true},
		{"str YES", "YES", true},
		{"str yes", " yes ", true},
		{"str y", "y", true},
		{"str on", "on", true},
		{"str 0", "0", false},
		{"str f", "f", false},
		{"str no", "No", false},
		{"str n", "n", false},
		{"str off", "off", false},
		{"bytes yes", []byte("yes"), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToBool(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}

func TestToBool_Error(t *testing.T) {
	for _, v := range []any{nil, "", "abc", struct{}{}} {
		if _, err := ToBool(v); err == nil {
			t.Fatalf("ToBool(%v) expected error", v)
		}
	}
}

// -------- ToString --------

func TestToString_Basic(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"string", "hi", "hi"},
		{"bytes", []byte("by"), "by"},
		{"bool", true, "true"},
		{"int", 42, "42"},
		{"int8", int8(-1), "-1"},
		{"int16", int16(16), "16"},
		{"int32", int32(32), "32"},
		{"int64", int64(64), "64"},
		{"uint", uint(10), "10"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"float32", float32(2.5), "2.5"},
		{"float64", 3.0, "3"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToString(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestToString_Error(t *testing.T) {
	for _, v := range []any{nil, struct{}{}, func() {}} {
		if _, err := ToString(v); err == nil {
			t.Fatalf("ToString(%v) expected error", v)
		}
	}
}

// -------- MustXxx --------

func TestMustXxx_Fallback(t *testing.T) {
	if MustInt(nil, -1) != -1 {
		t.Fatal("MustInt fallback")
	}
	if MustInt64(nil, -2) != -2 {
		t.Fatal("MustInt64 fallback")
	}
	if MustUint64(nil, 99) != 99 {
		t.Fatal("MustUint64 fallback")
	}
	if MustFloat64(nil, 3.14) != 3.14 {
		t.Fatal("MustFloat64 fallback")
	}
	if MustString(nil, "fallback") != "fallback" {
		t.Fatal("MustString fallback")
	}
	if MustBool(nil, true) != true {
		t.Fatal("MustBool fallback")
	}

	if MustInt(5, -1) != 5 {
		t.Fatal("MustInt success")
	}
	if MustString("ok", "fallback") != "ok" {
		t.Fatal("MustString success")
	}
}

// -------- ToDuration / MustDuration --------

func TestToDuration(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want time.Duration
	}{
		{"duration", 3 * time.Second, 3 * time.Second},
		{"int ns", int(100), 100},
		{"int64 ns", int64(100), 100},
		{"float", float64(2500 * time.Millisecond), 2500 * time.Millisecond},
		{"string pure num", "500", 500},
		{"string go dur", "1h30m", 90 * time.Minute},
		{"string float sec", "1.5", 1500 * time.Millisecond},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToDuration(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}

	for _, v := range []any{nil, "", "abc", struct{}{}} {
		if _, err := ToDuration(v); err == nil {
			t.Fatalf("ToDuration(%v) expected error", v)
		}
	}

	if MustDuration(nil, 5*time.Second) != 5*time.Second {
		t.Fatal("MustDuration fallback")
	}
}

// -------- ToTime / MustTime --------

func TestToTime(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	cases := []struct {
		name string
		in   any
		want time.Time
	}{
		{"time.Time", now, now},
		{"rfc3339", "2024-01-02T03:04:05Z", mustParseTime(t, "2006-01-02T15:04:05Z", "2024-01-02T03:04:05Z")},
		{"rfc3339nano", "2024-01-02T03:04:05.123456789Z", mustParseTime(t, time.RFC3339Nano, "2024-01-02T03:04:05.123456789Z")},
		{"simple date", "2024-01-02", mustParseTime(t, "2006-01-02", "2024-01-02")},
		{"datetime", "2024-01-02 03:04:05", mustParseTime(t, "2006-01-02 15:04:05", "2024-01-02 03:04:05")},
		{"unix second", int64(1700000000), time.Unix(1700000000, 0)},
		{"unix nanosecond", int64(1700000000000000000), time.Unix(0, 1700000000000000000)},
		{"uint unix", uint64(1700000000), time.Unix(1700000000, 0)},
		{"float unix", 1700000000.5, time.Unix(1700000000, 500000000)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToTime(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(c.want) {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}

	for _, v := range []any{nil, "", "abc", struct{}{}} {
		if _, err := ToTime(v); err == nil {
			t.Fatalf("ToTime(%v) expected error", v)
		}
	}

	fallback := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if !MustTime(nil, fallback).Equal(fallback) {
		t.Fatal("MustTime fallback")
	}
}

func mustParseTime(t *testing.T, layout, s string) time.Time {
	t.Helper()
	tt, err := time.Parse(layout, s)
	if err != nil {
		t.Fatal(err)
	}
	return tt
}

// -------- unixToTime 边界 --------

func TestUnixToTime_Magnitude(t *testing.T) {
	// 秒
	tm, err := unixToTime(1700000000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tm.Unix() != 1700000000 {
		t.Fatalf("second %d", tm.Unix())
	}

	// 毫秒
	tm, err = unixToTime(1700000000000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tm.Unix() != 1700000000 {
		t.Fatalf("ms unix got %d", tm.Unix())
	}

	// 微秒
	tm, err = unixToTime(1700000000000000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tm.Unix() != 1700000000 {
		t.Fatalf("us unix got %d", tm.Unix())
	}

	// 纳秒
	tm, err = unixToTime(1700000000000000000, nil)
	if err != nil {
		t.Fatal(err)
	}
	if tm.Unix() != 1700000000 {
		t.Fatalf("ns unix got %d", tm.Unix())
	}

	// 错误传播：直接传 err
	if _, err := unixToTime(0, errPlaceholder); err == nil {
		t.Fatal("expected error")
	}
}

// unixToTime 是私有函数，已通过 ToTime 的数值路径间接覆盖。

// -------- ToStringSlice --------

func TestToStringSlice(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want []string
	}{
		{"[]string", []string{"a", "b"}, []string{"a", "b"}},
		{"[]any", []any{1, "x", true}, []string{"1", "x", "true"}},
		{"csv", " a, b , ,c", []string{"a", "b", "c"}},
		{"empty csv", "", []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToStringSlice(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("len got %d want %d", len(got), len(c.want))
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("[%d] got %q want %q", i, got[i], c.want[i])
				}
			}
		})
	}

	for _, v := range []any{nil, 123, struct{}{}} {
		if _, err := ToStringSlice(v); err == nil {
			t.Fatalf("ToStringSlice(%v) expected error", v)
		}
	}

	// []any 内部 ToString 失败的路径
	if _, err := ToStringSlice([]any{func() {}}); err == nil {
		t.Fatal("expected error from bad element")
	}
}

// -------- ToIntSlice --------

func TestToIntSlice(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want []int
	}{
		{"csv", "1, 2, 3", []int{1, 2, 3}},
		{"[]string", []string{"4", "5"}, []int{4, 5}},
		{"[]int", []int{6, 7}, []int{6, 7}},
		{"[]any numeric", []any{8, "9", uint(10)}, []int{8, 9, 10}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToIntSlice(c.in)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("len got %d want %d", len(got), len(c.want))
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("[%d] got %d want %d", i, got[i], c.want[i])
				}
			}
		})
	}

	// 错误路径
	for _, v := range []any{"a,b,c", []string{"a", "b"}, []any{func() {}, 1}} {
		if _, err := ToIntSlice(v); err == nil {
			t.Fatalf("ToIntSlice(%v) expected error", v)
		}
	}
	if _, err := ToIntSlice(nil); err == nil {
		t.Fatal("nil expected error")
	}
	if _, err := ToIntSlice(123); err == nil {
		t.Fatal("scalar expected error")
	}
}

// -------- ToStringMap --------

func TestToStringMap(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{"map[string]any", map[string]any{"a": 1}},
		{"map[string]string", map[string]string{"b": "2"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := ToStringMap(c.in)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, v := range []any{nil, map[int]string{}, struct{}{}} {
		if _, err := ToStringMap(v); err == nil {
			t.Fatalf("ToStringMap(%v) expected error", v)
		}
	}
}

// -------- ToAddr --------

func TestToAddr(t *testing.T) {
	addrPort := netip.MustParseAddrPort("127.0.0.1:8080")
	addr := netip.MustParseAddr("127.0.0.1")

	cases := []struct {
		name    string
		in      any
		want    netip.AddrPort
		wantErr bool
	}{
		{"AddrPort", addrPort, addrPort, false},
		{"Addr", addr, netip.AddrPortFrom(addr, 0), false},
		{"string addrport", "10.0.0.1:9000", netip.MustParseAddrPort("10.0.0.1:9000"), false},
		{"string addr only", "10.0.0.1", netip.AddrPortFrom(netip.MustParseAddr("10.0.0.1"), 0), false},
		{"bad string", "not-an-addr", netip.AddrPort{}, true},
		{"struct", struct{}{}, netip.AddrPort{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToAddr(c.in)
			if (err != nil) != c.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, c.wantErr)
			}
			if !c.wantErr && got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}
