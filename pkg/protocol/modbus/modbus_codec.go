// Package modbus 内部 —— 值编解码层（纯 bytes + DataType，无 scale/offset）
//
// 只负责 bytes ↔ 语义值 的可逆转换。
// scale / offset / enum / precision 在上层处理（mapper.DecodeResponse / Write 调用方）。
package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Decode 把 Modbus 响应原始字节解码成语义值
//
// raw      原始 big-endian 字节（holding/input 是 2*n 字节）
// dt       DataType enum
// little   true=每个 16-bit word 内部小端序
// swap     true=32-bit 高 word / 低 word 交换
func Decode(raw []byte, dt DataType, little, swap bool) (any, error) {
	if len(raw) < 1 {
		return nil, fmt.Errorf("modbus Decode: 空数据")
	}
	switch dt {
	case TypeUint8:
		if len(raw) < 1 {
			return nil, fmt.Errorf("modbus Decode uint8: 需 1 字节")
		}
		return raw[0], nil

	case TypeInt8:
		if len(raw) < 1 {
			return nil, fmt.Errorf("modbus Decode int8: 需 1 字节")
		}
		return int8(raw[0]), nil

	case TypeUint16:
		if len(raw) < 2 {
			return nil, fmt.Errorf("modbus Decode uint16: 需 2 字节，得 %d", len(raw))
		}
		if little {
			return uint16(raw[0]) | uint16(raw[1])<<8, nil
		}
		return binary.BigEndian.Uint16(raw[:2]), nil

	case TypeInt16:
		if len(raw) < 2 {
			return nil, fmt.Errorf("modbus Decode int16: 需 2 字节，得 %d", len(raw))
		}
		var v uint16
		if little {
			v = uint16(raw[0]) | uint16(raw[1])<<8
		} else {
			v = binary.BigEndian.Uint16(raw[:2])
		}
		return int16(v), nil

	case TypeUint32:
		if len(raw) < 4 {
			return nil, fmt.Errorf("modbus Decode uint32: 需 4 字节，得 %d", len(raw))
		}
		var r0, r1 uint16
		if little {
			r0 = uint16(raw[0]) | uint16(raw[1])<<8
			r1 = uint16(raw[2]) | uint16(raw[3])<<8
		} else {
			r0 = binary.BigEndian.Uint16(raw[0:2])
			r1 = binary.BigEndian.Uint16(raw[2:4])
		}
		if swap {
			return uint32(r1)<<16 | uint32(r0), nil
		}
		return uint32(r0)<<16 | uint32(r1), nil

	case TypeInt32:
		v, err := Decode(raw, TypeUint32, little, swap)
		if err != nil {
			return nil, err
		}
		u32, ok := v.(uint32)
		if !ok {
			return nil, fmt.Errorf("modbus Decode int32: 内部类型断言失败")
		}
		return int32(u32), nil

	case TypeFloat32:
		if len(raw) < 4 {
			return nil, fmt.Errorf("modbus Decode float32: 需 4 字节，得 %d", len(raw))
		}
		var r0, r1 uint16
		if little {
			r0 = uint16(raw[0]) | uint16(raw[1])<<8
			r1 = uint16(raw[2]) | uint16(raw[3])<<8
		} else {
			r0 = binary.BigEndian.Uint16(raw[0:2])
			r1 = binary.BigEndian.Uint16(raw[2:4])
		}
		if swap {
			r0, r1 = r1, r0
		}
		buf := make([]byte, 4)
		binary.BigEndian.PutUint16(buf[0:2], r0)
		binary.BigEndian.PutUint16(buf[2:4], r1)
		return float64(math.Float32frombits(binary.BigEndian.Uint32(buf))), nil

	case TypeFloat64:
		if len(raw) < 8 {
			return nil, fmt.Errorf("modbus Decode float64: 需 8 字节，得 %d", len(raw))
		}
		// 与 float32 保持一致的字节序分层语义：
		//   little → 每个 16-bit word 内部字节交换
		//   swap   → word 顺序反转（w0 w1 w2 w3 → w3 w2 w1 w0）
		var b [8]byte
		copy(b[:], raw[:8])
		if little {
			for i := 0; i < 8; i += 2 {
				b[i], b[i+1] = b[i+1], b[i]
			}
		}
		if swap {
			b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7] =
				b[6], b[7], b[4], b[5], b[2], b[3], b[0], b[1]
		}
		bits := binary.BigEndian.Uint64(b[:])
		return math.Float64frombits(bits), nil

	case TypeBool:
		if raw[0]&0x01 != 0 {
			return true, nil
		}
		return false, nil

	default:
		return nil, fmt.Errorf("modbus Decode: 未知 DataType %v", dt)
	}
}

// Encode 把语义值按 DataType + ByteOrder 编码成原始字节
func Encode(value any, dt DataType, little, swap bool) ([]byte, error) {
	switch dt {
	case TypeUint8:
		n, ok := toUint16(value)
		if !ok {
			return nil, fmt.Errorf("modbus Encode uint8: 无法转换 %T", value)
		}
		return []byte{byte(n)}, nil

	case TypeInt8:
		switch v := value.(type) {
		case int8:
			return []byte{byte(v)}, nil
		case int:
			return []byte{byte(int8(v))}, nil
		case float64:
			return []byte{byte(int8(v))}, nil
		default:
			return nil, fmt.Errorf("modbus Encode int8: 无法转换 %T", value)
		}

	case TypeUint16:
		n, ok := toUint16(value)
		if !ok {
			return nil, fmt.Errorf("modbus Encode uint16: 无法转换 %T", value)
		}
		buf := make([]byte, 2)
		if little {
			buf[0] = byte(n & 0xFF)
			buf[1] = byte(n >> 8)
		} else {
			binary.BigEndian.PutUint16(buf, n)
		}
		return buf, nil

	case TypeInt16:
		var n uint16
		switch v := value.(type) {
		case int16:
			n = uint16(v)
		case int:
			n = uint16(int16(v))
		case float64:
			n = uint16(int16(v))
		default:
			return nil, fmt.Errorf("modbus Encode int16: 无法转换 %T", value)
		}
		buf := make([]byte, 2)
		if little {
			buf[0] = byte(n & 0xFF)
			buf[1] = byte(n >> 8)
		} else {
			binary.BigEndian.PutUint16(buf, n)
		}
		return buf, nil

	case TypeUint32:
		var u uint32
		switch v := value.(type) {
		case uint32:
			u = v
		case int:
			u = uint32(v)
		case float64:
			u = uint32(v)
		default:
			return nil, fmt.Errorf("modbus Encode uint32: 无法转换 %T", value)
		}
		r0 := uint16(u >> 16)
		r1 := uint16(u & 0xFFFF)
		if swap {
			r0, r1 = r1, r0
		}
		buf := make([]byte, 4)
		if little {
			buf[0] = byte(r0 & 0xFF)
			buf[1] = byte(r0 >> 8)
			buf[2] = byte(r1 & 0xFF)
			buf[3] = byte(r1 >> 8)
		} else {
			binary.BigEndian.PutUint16(buf[0:2], r0)
			binary.BigEndian.PutUint16(buf[2:4], r1)
		}
		return buf, nil

	case TypeInt32:
		var i int32
		switch v := value.(type) {
		case int32:
			i = v
		case int:
			i = int32(v)
		case float64:
			i = int32(v)
		default:
			return nil, fmt.Errorf("modbus Encode int32: 无法转换 %T", value)
		}
		return Encode(uint32(i), TypeUint32, little, swap)

	case TypeFloat32:
		var f float32
		switch v := value.(type) {
		case float32:
			f = v
		case float64:
			f = float32(v)
		case int:
			f = float32(v)
		default:
			return nil, fmt.Errorf("modbus Encode float32: 无法转换 %T", value)
		}
		bits := math.Float32bits(f)
		r0 := uint16(bits >> 16)
		r1 := uint16(bits & 0xFFFF)
		if swap {
			r0, r1 = r1, r0
		}
		buf := make([]byte, 4)
		if little {
			buf[0] = byte(r0 & 0xFF)
			buf[1] = byte(r0 >> 8)
			buf[2] = byte(r1 & 0xFF)
			buf[3] = byte(r1 >> 8)
		} else {
			binary.BigEndian.PutUint16(buf[0:2], r0)
			binary.BigEndian.PutUint16(buf[2:4], r1)
		}
		return buf, nil

	case TypeFloat64:
		var f float64
		switch v := value.(type) {
		case float64:
			f = v
		case float32:
			f = float64(v)
		case int:
			f = float64(v)
		default:
			return nil, fmt.Errorf("modbus Encode float64: 无法转换 %T", value)
		}
		// 与 Decode 严格互逆：Decode 是「先 little 再 swap」，故此处为「先 swap 再 little」
		var b [8]byte
		bits := math.Float64bits(f)
		binary.BigEndian.PutUint64(b[:], bits)
		if swap {
			b = [8]byte{b[6], b[7], b[4], b[5], b[2], b[3], b[0], b[1]}
		}
		if little {
			b = [8]byte{b[1], b[0], b[3], b[2], b[5], b[4], b[7], b[6]}
		}
		return b[:], nil

	case TypeBool:
		b, ok := toUint16(value)
		if !ok {
			return nil, fmt.Errorf("modbus Encode bool: 无法转换 %T", value)
		}
		if b != 0 {
			return []byte{0xFF, 0x00}, nil
		}
		return []byte{0x00, 0x00}, nil

	default:
		return nil, fmt.Errorf("modbus Encode: 未知 DataType %v", dt)
	}
}

// toUint16 把 any 转 uint16，返回 (v, ok) —— codec 内部工具
func toUint16(v any) (uint16, bool) {
	switch x := v.(type) {
	case int:
		return uint16(x), true
	case int16:
		return uint16(x), true
	case uint16:
		return x, true
	case int32:
		return uint16(x), true
	case uint32:
		return uint16(x), true
	case float64:
		return uint16(x), true
	case float32:
		return uint16(x), true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

// ============ 上层 scale/offset/precision 辅助（Decode/Encode 不处理这些）============

// applyScaleOffsetPrecision 把 Decode 的原始结果套 scale/offset/precision
// 输入 any 来自 Decode（uint16/int16/uint32/int32/float64）→ 转 float64 → 套公式 → 四舍五入
func applyScaleOffsetPrecision(val any, scale, offset float64, precision int) float64 {
	v := toFloat64(val)
	if scale == 0 {
		scale = 1
	}
	result := v*scale + offset
	if precision > 0 {
		mul := 1.0
		for i := 0; i < precision; i++ {
			mul *= 10
		}
		result = math.Round(result*mul) / mul
	}
	return result
}

// applyBitResult 处理 coil/discrete/bit 抽位 —— 默认返回 bool，显式配了 scale/offset/precision 转数值
func applyBitResult(bit bool, scale, offset float64, precision int) any {
	if (scale == 0 || scale == 1) && offset == 0 && precision <= 0 {
		return bit
	}
	v := float64(0)
	if bit {
		v = 1
	}
	return applyScaleOffsetPrecision(v, scale, offset, precision)
}

func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int16:
		return float64(x)
	case uint16:
		return float64(x)
	case int32:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// ============ 逆向枚举（Write 路径）============

// reverseEnum 逆向枚举 —— 把语义值（string/int/bool）还原成枚举原始 int value
func reverseEnum(semantic any, enums []EnumDef) (int, error) {
	if len(enums) == 0 {
		switch v := semantic.(type) {
		case bool:
			if v {
				return 1, nil
			}
			return 0, nil
		case float64:
			return int(v), nil
		case int:
			return v, nil
		case string:
			// 未定义枚举时无法把字符串映射成寄存器值 —— 报错而不是静默写 0
			return 0, fmt.Errorf("modbus: 未定义枚举，无法把字符串 %q 转成写入值", v)
		default:
			return 0, fmt.Errorf("modbus: 无法把 %T 转 int 写入", semantic)
		}
	}

	switch v := semantic.(type) {
	case bool:
		target := 0
		if v {
			target = 1
		}
		for _, e := range enums {
			if e.Value == target {
				return target, nil
			}
		}
		return target, nil
	case string:
		for _, e := range enums {
			if e.Key == v {
				return e.Value, nil
			}
		}
		for _, e := range enums {
			if e.Name == v {
				return e.Value, nil
			}
		}
		return 0, fmt.Errorf("modbus: 枚举值 %q 未定义", v)
	case float64:
		target := int(v)
		for _, e := range enums {
			if e.Value == target {
				return target, nil
			}
		}
		return target, nil
	case int:
		for _, e := range enums {
			if e.Value == v {
				return v, nil
			}
		}
		return v, nil
	default:
		return 0, fmt.Errorf("modbus: 无法把 %T 逆向枚举", semantic)
	}
}
