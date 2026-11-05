// Package meter 从 meter.lua 移植的通用表计数据解码库
//
// 被 DLT645、CJT188 等协议共享。支持：
//   - 16 种 BCD 格式（含内置 rate）
//   - hex / u8 / u16 整数
//   - datetime / date 字符串
//   - CJT188 单位转换
package meter

import (
	"fmt"
	"math"
	"strings"
)

// Format meter.lua types 表移植
type Format struct {
	Typ  string  // "bcd" / "hex" / "u8" / "u16" / "date" / "datetime" / "datetime5" / "datetime6"
	Size int     // 字节数
	Rate float64 // 内置 rate（BCD 小数位）
}

// Formats meter.lua 完整类型表
var Formats = map[string]Format{
	"XXXXXXXX":        {Typ: "bcd", Size: 4},
	"XXXXXX":          {Typ: "bcd", Size: 3},
	"XXXXXX.XX":       {Typ: "bcd", Size: 4, Rate: 0.01},
	"XXXX":            {Typ: "bcd", Size: 2},
	"XXXX.XX":         {Typ: "bcd", Size: 3, Rate: 0.01},
	"XXXX.XXXX":       {Typ: "bcd", Size: 4, Rate: 0.0001},
	"XXX.X":           {Typ: "bcd", Size: 2, Rate: 0.1},
	"XXX.XXX":         {Typ: "bcd", Size: 3, Rate: 0.001},
	"XX":              {Typ: "bcd", Size: 1},
	"XX.XX":           {Typ: "bcd", Size: 2, Rate: 0.01},
	"XX.XXXX":         {Typ: "bcd", Size: 3, Rate: 0.0001},
	"X.XXX":           {Typ: "bcd", Size: 2, Rate: 0.001},
	"HH":              {Typ: "hex", Size: 1},
	"HHHH":            {Typ: "hex", Size: 2},
	"YYYYMMDDhhmmss":  {Typ: "datetime", Size: 7},
	"YYMMDDhhmmss":    {Typ: "datetime6", Size: 6},
	"YYMMDDhhmm":      {Typ: "datetime5", Size: 5},
	"YYYYMMDD":        {Typ: "date", Size: 4},
	"uint8":           {Typ: "u8", Size: 1},
	"uint16":          {Typ: "u16", Size: 2},
	"int8":            {Typ: "u8", Size: 1},
	"int16":           {Typ: "u16", Size: 2},
	"uint32":          {Typ: "u16", Size: 4},
	"int32":           {Typ: "u16", Size: 4},
	"float32":         {Typ: "bcd", Size: 4},
	"float64":         {Typ: "bcd", Size: 8},
	"string":          {Typ: "hex", Size: 4},
	"bcd":             {Typ: "bcd", Size: 4},
	"hex":             {Typ: "hex", Size: 2},
}

// FormatSize 返回 type 的字节长度（未知则 4）
func FormatSize(typ string) (int, bool) {
	if f, ok := Formats[typ]; ok {
		return f.Size, true
	}
	return 4, false
}

// Decode 对齐 meter.lua decode(data, type, reverse)
// 额外接收 scale/offset/precision 做语义处理
func Decode(data []byte, typ string, reverse bool, scale, offset float64, precision int) (any, error) {
	f, ok := Formats[typ]
	if !ok {
		return nil, fmt.Errorf("meter: 未知数据类型 %q", typ)
	}
	if len(data) < f.Size {
		return nil, fmt.Errorf("meter: 数据长度不足 need=%d have=%d", f.Size, len(data))
	}

	b := make([]byte, f.Size)
	copy(b, data[:f.Size])
	if reverse && f.Size > 1 {
		for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
			b[i], b[j] = b[j], b[i]
		}
	}

	switch f.Typ {
	case "bcd":
		raw := decodeBCD(b)
		value := float64(raw)
		if f.Rate != 0 {
			value = value * f.Rate
		}
		if scale != 0 && scale != 1 {
			value = value * scale
		}
		value = value + offset
		return ApplyPrecision(value, precision), nil

	case "hex":
		var v uint64
		for _, by := range b {
			v = (v << 8) | uint64(by)
		}
		value := float64(v)*scale + offset
		return ApplyPrecision(value, precision), nil

	case "u8":
		value := float64(b[0])*scale + offset
		return ApplyPrecision(value, precision), nil

	case "u16":
		var v uint64
		for _, by := range b {
			v = (v << 8) | uint64(by)
		}
		value := float64(v)*scale + offset
		return ApplyPrecision(value, precision), nil

	case "date":
		return HexString(b), nil
	case "datetime":
		return HexString(b), nil
	case "datetime6":
		return "20" + HexString(b), nil
	case "datetime5":
		return "20" + HexString(b) + "00", nil

	default:
		return nil, fmt.Errorf("meter: 不支持的格式 %q", f.Typ)
	}
}

// decodeBCD BCD 解码：每个高 nibble*10 + 低 nibble
func decodeBCD(b []byte) uint64 {
	var v uint64
	for _, by := range b {
		v = v*100 + uint64(((by>>4)&0x0F)*10+(by&0x0F))
	}
	return v
}

// HexString 字节转大写 hex 字符串
func HexString(b []byte) string {
	var sb strings.Builder
	for _, by := range b {
		sb.WriteString(fmt.Sprintf("%02X", by))
	}
	return sb.String()
}

// ApplyPrecision 按 precision 保留小数
func ApplyPrecision(v float64, p int) float64 {
	if p <= 0 {
		return v
	}
	pow := 1.0
	for i := 0; i < p; i++ {
		pow *= 10
	}
	return float64(int64(v*pow+0.5)) / pow
}

// ============================================================
// CJT188 单位转换（unit_convert 移植）
// ============================================================

// UnitConvert 对齐 cjt188.lua unit_convert
// 返回统一单位下的值：
//   0x02-0x0A → kWh; 0x0B-0x13 → kJ; 0x14-0x1C → kW;
//   0x29-0x2E → m3; 0x32-0x37 → m3/h; 0x40-0x4B → kJ/h
func UnitConvert(unit byte, value float64) float64 {
	switch {
	case unit == 0x01: // J → kJ
		return value * 0.001
	case unit >= 0x02 && unit <= 0x0A: // Wh → kWh
		return value * math.Pow(10, float64(unit-0x05))
	case unit >= 0x0B && unit <= 0x13: // kJ 系列
		return value * math.Pow(10, float64(unit-0x0B))
	case unit >= 0x14 && unit <= 0x1C: // W 系列 → kW
		return value * math.Pow(10, float64(unit-0x17))
	case unit >= 0x29 && unit <= 0x2E: // L/m3 → m3
		return value * math.Pow(10, float64(unit-0x2C))
	case unit >= 0x32 && unit <= 0x37: // L/h/m3/h → m3/h
		return value * math.Pow(10, float64(unit-0x35))
	case unit >= 0x40 && unit <= 0x4B: // J/h → kJ/h
		return value * math.Pow(10, float64(unit-0x43))
	}
	return value
}
