// Package modbus 实现 Modbus TCP / RTU 两种协议的 protocol.Protocol 接口
package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/god-jason/iot-master/pkg/protocol"
)

// ============================================================
// Modbus 子包私有公共类型（协议自定义，不依赖公共 interface）
// ============================================================

// Frame 协议无关的帧（modbus 子包内部编解码共用）
//
// Data 是响应 PDU（不含 FC，首字节为 byteCount），是值的唯一真源；
// 需要寄存器/位值列表时用 Uint16s() 派生 —— 不另存一份，避免两者不一致。
type Frame struct {
	Protocol string

	Function  byte
	StartAddr uint16
	Quantity  uint16
	SlaveAddr byte
	UnitID    byte

	Data []byte
	Raw  []byte

	IsError  bool
	ErrorMsg string
}

// Uint16s 从 Data 派生值列表：线圈/离散按位展开，寄存器按字展开。
// 解析失败（如异常响应、数据不完整）返回 nil。
func (f *Frame) Uint16s() []uint16 {
	vals, err := ParseReadResponse(f.Function, f.Data)
	if err != nil {
		return nil
	}
	return vals
}

// MapItem 协议无关的纯点位数据结构
// Parse 阶段会把 FC/StartAddr/Quantity 算好填进来（纯计算结果，不来自输入）
type MapItem struct {
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	DataType  DataType  `json:"data_type"`
	LittleEnd bool      `json:"little_endian,omitempty"`
	Swap      bool      `json:"swap,omitempty"`
	Scale     float64   `json:"scale,omitempty"`
	Offset    float64   `json:"offset,omitempty"`
	Precision int       `json:"precision,omitempty"`
	Bit       int       `json:"bit,omitempty"`
	Enums     []EnumDef `json:"enums,omitempty"`

	// Parse 阶段计算填充（json tag 为 "-"，不参与序列化）
	FC        byte   `json:"-"`
	StartAddr uint16 `json:"-"`
	Quantity  uint16 `json:"-"`
}

// DataType 数据类型枚举 —— Register 阶段从 string 解析成数值存进 MapItem
type DataType int

const (
	TypeUnknown DataType = iota
	TypeUint8            // uint8 / byte / u8
	TypeInt8             // int8 / char / s8
	TypeUint16           // uint16 / word / u16
	TypeInt16            // int16 / short / s16
	TypeUint32           // uint32 / dword / u32
	TypeInt32            // int32 / long / s32
	TypeFloat32          // float32
	TypeFloat64          // float64 / double
	TypeBool
	TypeBCD
)

// dataTypeNames 枚举名（反向查找用）
var dataTypeNames = map[string]DataType{
	"uint8": TypeUint8, "byte": TypeUint8, "u8": TypeUint8, "ubyte": TypeUint8,
	"int8": TypeInt8, "char": TypeInt8, "s8": TypeInt8,
	"uint16": TypeUint16, "word": TypeUint16, "u16": TypeUint16,
	"int16": TypeInt16, "short": TypeInt16, "s16": TypeInt16,
	"uint32": TypeUint32, "dword": TypeUint32, "u32": TypeUint32,
	"int32": TypeInt32, "long": TypeInt32, "s32": TypeInt32,
	"float32": TypeFloat32, "float": TypeFloat32,
	"float64": TypeFloat64, "double": TypeFloat64,
	"bool": TypeBool,
	"bcd":  TypeBCD,
}

// dataTypeQuantity 每种类型占多少个 holding/input 寄存器
var dataTypeQuantity = map[DataType]uint16{
	TypeUint8: 1, TypeInt8: 1, TypeBool: 1, TypeBCD: 1,
	TypeUint16: 1, TypeInt16: 1,
	TypeUint32: 2, TypeInt32: 2, TypeFloat32: 2,
	TypeFloat64: 4,
}

// parseDataType 把 string 解析成 DataType enum（Register 阶段调用）
func parseDataType(s string) DataType {
	if s == "" {
		return TypeUint16
	}
	if v, ok := dataTypeNames[strings.ToLower(strings.TrimSpace(s))]; ok {
		return v
	}
	return TypeUnknown
}

// registerQuantity 返回该 DataType 在 holding/input 里占多少寄存器
func registerQuantity(d DataType) uint16 {
	if q, ok := dataTypeQuantity[d]; ok {
		return q
	}
	return 1
}

// EnumDef 枚举值（与 thing.Enum 兼容但 modbus 自有）
type EnumDef struct {
	Key   string `json:"key,omitempty"`
	Name  string `json:"name,omitempty"`
	Value int    `json:"value,omitempty"`
}

// mapToEnum 把原始值（bool/int/float64）映射到枚举定义中的 Key/Name
// 优先精确匹配 enum.Value；如果没有匹配则返回原值 + ok=false
func mapToEnum(raw any, enums []EnumDef) (any, bool) {
	var rawNum int
	switch v := raw.(type) {
	case bool:
		if v {
			rawNum = 1
		} else {
			rawNum = 0
		}
	case float64:
		// 用 Round 而非直接截断：浮点解码结果常带微小误差（如 2.9999999），
		// 直接截断会错配枚举。
		rawNum = int(math.Round(v))
	case int:
		rawNum = v
	case uint16:
		rawNum = int(v)
	default:
		return raw, false
	}

	for _, e := range enums {
		if e.Value == rawNum {
			if e.Key != "" {
				return e.Key, true
			}
			if e.Name != "" {
				return e.Name, true
			}
			return rawNum, true
		}
	}
	return raw, false
}

// Function Code
const (
	FCReadCoils            byte = 0x01
	FCReadDiscreteInputs   byte = 0x02
	FCReadHoldingRegisters byte = 0x03
	FCReadInputRegisters   byte = 0x04
	FCWriteSingleCoil      byte = 0x05
	FCWriteSingleRegister  byte = 0x06
	FCWriteMultipleCoils   byte = 0x0F
	FCWriteMultipleRegs    byte = 0x10
)

var (
	ErrCRC             = protocol.ErrCRC
	ErrNeedMore        = protocol.ErrNeedMore
	ErrInvalidFrame    = protocol.ErrInvalidFrame
	ErrIllegalFunction = errors.New("modbus: 非法功能码")
)

// ============ CRC-16 (Modbus) ============

func crc16Modbus(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// ============ PDU 辅助 ============
//
// 本包统一约定：PDU 不含 FC（FC 由帧头或单独参数携带），
// 这样 TCP 的 MBAP 和 RTU 的帧编码可以共享同一份 PDU 构造逻辑。

// BuildReadRequest 构造读请求 PDU（不含 FC）：addr(2) + quantity(2)
func BuildReadRequest(addr, quantity uint16) []byte {
	pdu := make([]byte, 4)
	binary.BigEndian.PutUint16(pdu[0:2], addr)
	binary.BigEndian.PutUint16(pdu[2:4], quantity)
	return pdu
}

// ParseReadResponse 解析读响应 PDU（不含 FC）为语义值列表。
// fc 仅用于区分线圈/离散（按位展开）与寄存器（按字展开）。
func ParseReadResponse(fc byte, pdu []byte) ([]uint16, error) {
	if len(pdu) < 1 {
		return nil, ErrNeedMore
	}
	if fc&0x80 != 0 {
		exc := pdu[0]
		return nil, fmt.Errorf("modbus exception: %02X", exc)
	}
	byteCount := int(pdu[0])
	data := pdu[1:]
	if len(data) < byteCount {
		return nil, ErrNeedMore
	}
	data = data[:byteCount]

	switch fc {
	case FCReadCoils, FCReadDiscreteInputs:
		values := make([]uint16, byteCount*8)
		for i, b := range data {
			for j := 0; j < 8; j++ {
				if b&(1<<j) != 0 {
					values[i*8+j] = 1
				}
			}
		}
		return values, nil
	case FCReadHoldingRegisters, FCReadInputRegisters:
		values := make([]uint16, 0, byteCount/2)
		for i := 0; i+1 < len(data); i += 2 {
			values = append(values, binary.BigEndian.Uint16(data[i:i+2]))
		}
		return values, nil
	default:
		return nil, ErrIllegalFunction
	}
}

// BuildFrameFromPdu 把解码后的帧（pdu 不含 FC）组装成 Frame
func BuildFrameFromPdu(proto string, slaveAddr, unitID, fc byte, pdu, raw []byte) *Frame {
	df := &Frame{
		Protocol:  proto,
		SlaveAddr: slaveAddr,
		UnitID:    unitID,
		Function:  fc,
		Data:      append([]byte(nil), pdu...),
		Raw:       append([]byte(nil), raw...),
	}
	if fc&0x80 != 0 {
		df.IsError = true
		if len(pdu) > 0 {
			df.ErrorMsg = fmt.Sprintf("modbus exception fc=%02X exc=%02X", fc, pdu[0])
		}
		return df
	}
	if fc == FCWriteSingleRegister || fc == FCWriteMultipleRegs || fc == FCWriteSingleCoil {
		return df
	}
	// 仅做可解析性校验（异常/非法数据在此标记 IsError），值不再另存一份
	if _, err := ParseReadResponse(fc, pdu); err != nil {
		df.IsError = true
		df.ErrorMsg = err.Error()
	}
	return df
}
