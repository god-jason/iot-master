package protocol

import (
	"errors"
	"fmt"

	"github.com/god-jason/iot-master/bin"
	"github.com/god-jason/iot-master/product"
	"github.com/spf13/cast"
)

type Point interface {
	Encode(data any) ([]byte, error)
	Parse(address uint16, buf []byte) (any, error)
}

type PointBit struct {
	product.Point //继承

	Address uint16 `json:"address"` //偏移
}

func (p *PointBit) Encode(data any) ([]byte, error) {
	val := cast.ToBool(data)
	if val {
		return []byte{0xFF, 00}, nil
	} else {
		return []byte{0x00, 00}, nil
	}
}

func (p *PointBit) Parse(address uint16, buf []byte) (any, error) {
	l := len(buf)
	offset := int(p.Address - address)
	cur := offset / 8
	bit := offset % 8

	if cur >= l {
		return nil, errors.New("长度不够")
	}

	ret := buf[cur] & (1 << bit)

	return ret > 0, nil
}

type Enumeration struct {
	Index uint   `json:"index"`           //索引
	Value string `json:"value"`           //枚举值
	Label string `json:"label,omitempty"` //显示
}

type PointWord struct {
	product.Point //继承

	Address      uint16  `json:"address"`           //偏移
	LittleEndian bool    `json:"le,omitempty"`      //小端模式
	Rate         float64 `json:"rate,omitempty"`    //倍率
	Correct      float64 `json:"correct,omitempty"` //纠正
	Bits         []*Bit  `json:"bits,omitempty"`    //位，1 2 3...

	Enumerations []*Enumeration `json:"enumerations,omitempty"` //枚举
}

type Bit struct {
	Name  string `json:"name"`            //名称
	Label string `json:"label,omitempty"` //显示
	Bit   int    `json:"bit"`             //偏移
}

func (p *PointWord) getEnumValue(index uint) (string, error) {
	for _, e := range p.Enumerations {
		if e.Index == index {
			return e.Value, nil
		}
	}
	return "", errors.New("找不到对应的枚举值")
}

func (p *PointWord) getEnumIndex(value string) (uint, error) {
	for _, e := range p.Enumerations {
		if e.Value == value {
			return e.Index, nil
		}
	}
	return 0, errors.New("找不到对应的枚举值")
}

func (p *PointWord) Encode(data any) ([]byte, error) {
	var ret []byte

	//枚举值
	if len(p.Enumerations) > 0 {
		var err error
		data, err = p.getEnumIndex(cast.ToString(data))
		if err != nil {
			return nil, err
		}
	}

	//纠正
	if p.Correct != 0 {
		data = cast.ToFloat64(data) - p.Correct
	}

	//倍率逆转换
	if p.Rate != 0 && p.Rate != 1 {
		data = cast.ToFloat64(data) / p.Rate
	}

	switch p.Type {
	case "bool", "boolean":
		val, err := cast.ToBoolE(data)
		var v uint16 = 0
		if val {
			v = 1
		}
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 2)
		if p.LittleEndian {
			bin.WriteUint16LittleEndian(ret, v)
		} else {
			bin.WriteUint16(ret, v)
		}
	case "int8":
		val, err := cast.ToInt8E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 2)
		if p.LittleEndian {
			bin.WriteUint16LittleEndian(ret, uint16(val))
		} else {
			bin.WriteUint16(ret, uint16(val))
		}
	case "uint8":
		val, err := cast.ToUint8E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 2)
		if p.LittleEndian {
			bin.WriteUint16LittleEndian(ret, uint16(val))
		} else {
			bin.WriteUint16(ret, uint16(val))
		}
	case "short", "int16":
		val, err := cast.ToInt16E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 2)
		if p.LittleEndian {
			bin.WriteUint16LittleEndian(ret, uint16(val))
		} else {
			bin.WriteUint16(ret, uint16(val))
		}
	case "word", "uint16":
		val, err := cast.ToUint16E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 2)
		if p.LittleEndian {
			bin.WriteUint16LittleEndian(ret, val)
		} else {
			bin.WriteUint16(ret, val)
		}
	case "int32", "int":
		val, err := cast.ToInt32E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 4)
		if p.LittleEndian {
			bin.WriteUint32LittleEndian(ret, uint32(val))
		} else {
			bin.WriteUint32(ret, uint32(val))
		}
	case "qword", "uint32", "uint":
		val, err := cast.ToUint32E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 4)
		if p.LittleEndian {
			bin.WriteUint32LittleEndian(ret, val)
		} else {
			bin.WriteUint32(ret, val)
		}
	case "int64":
		val, err := cast.ToInt64E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 8)
		if p.LittleEndian {
			bin.WriteUint64LittleEndian(ret, uint64(val))
		} else {
			bin.WriteUint64(ret, uint64(val))
		}
	case "uint64":
		val, err := cast.ToUint64E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 4)
		if p.LittleEndian {
			bin.WriteUint64LittleEndian(ret, val)
		} else {
			bin.WriteUint64(ret, val)
		}
	case "float", "float32":
		val, err := cast.ToFloat32E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 4)
		if p.LittleEndian {
			bin.WriteFloat32LittleEndian(ret, val)
		} else {
			bin.WriteFloat32(ret, val)
		}
	case "double", "float64":
		val, err := cast.ToFloat64E(data)
		if err != nil {
			return nil, err
		}
		ret = make([]byte, 8)
		if p.LittleEndian {
			bin.WriteFloat64LittleEndian(ret, val)
		} else {
			bin.WriteFloat64(ret, val)
		}
		//case "string", "bytes":
		//	switch v := data.(type) {
		//	case []byte:
		//		ret = v
		//	case string:
		//		ret = []byte(v)
		//	default:
		//		return nil, errors.New("string类型错误")
		//	}
		//case "hex":
		//	buf, err := hex.DecodeString(data.(string))
		//	if err != nil {
		//		return nil, err
		//	}
		//	ret = buf
	}

	return ret, nil
}

func (p *PointWord) Parse(address uint16, buf []byte) (any, error) {
	l := len(buf)

	offset := int((p.Address - address) * 2)
	//offset := p.Offset << 1
	if offset >= l {
		return nil, errors.New("长度不够")
	}

	var ret any
	switch p.Type {
	case "bool", "boolean":
		if len(buf[offset:]) < 2 {
			return nil, fmt.Errorf("bool长度不足2:%d", l)
		}
		if p.LittleEndian {
			ret = bin.ParseUint16LittleEndian(buf[offset:]) > 0
		} else {
			ret = bin.ParseUint16(buf[offset:]) > 0
		}
	case "int8":
		if len(buf[offset:]) < 2 {
			return nil, fmt.Errorf("int8长度不足2:%d", l)
		}
		if p.LittleEndian {
			ret = int8(buf[offset])
		} else {
			ret = int8(buf[offset+1])
		}
	case "uint8":
		if len(buf[offset:]) < 2 {
			return nil, fmt.Errorf("uint8长度不足2:%d", l)
		}
		if p.LittleEndian {
			ret = uint8(buf[offset])
		} else {
			ret = uint8(buf[offset+1])
		}
	case "short", "int16":
		if len(buf[offset:]) < 2 {
			return nil, fmt.Errorf("int16长度不足2:%d", l)
		}
		if p.LittleEndian {
			ret = int16(bin.ParseUint16LittleEndian(buf[offset:]))
		} else {
			ret = int16(bin.ParseUint16(buf[offset:]))
		}
	case "word", "uint16":
		if len(buf[offset:]) < 2 {
			return nil, fmt.Errorf("uint16长度不足2:%d", l)
		}
		if p.LittleEndian {
			ret = bin.ParseUint16LittleEndian(buf[offset:])
		} else {
			ret = bin.ParseUint16(buf[offset:])
		}
		//取位
		if p.Bits != nil && len(p.Bits) > 0 {
			rets := make(map[string]bool)
			for _, b := range p.Bits {
				rets[b.Name] = (ret.(uint16))&(1<<b.Bit) > 0
			}
			return rets, nil
		}
	case "int32", "int":
		if len(buf[offset:]) < 4 {
			return nil, fmt.Errorf("int32长度不足4:%d", l)
		}
		if p.LittleEndian {
			ret = int32(bin.ParseUint32LittleEndian(buf[offset:]))
		} else {
			ret = int32(bin.ParseUint32(buf[offset:]))
		}
	case "qword", "uint32", "uint":
		if len(buf[offset:]) < 4 {
			return nil, fmt.Errorf("uint32长度不足4:%d", l)
		}
		if p.LittleEndian {
			ret = bin.ParseUint32LittleEndian(buf[offset:])
		} else {
			ret = bin.ParseUint32(buf[offset:])
		}
	case "int64":
		if len(buf[offset:]) < 8 {
			return nil, fmt.Errorf("int64长度不足8:%d", l)
		}
		if p.LittleEndian {
			ret = int64(bin.ParseUint64LittleEndian(buf[offset:]))
		} else {
			ret = int64(bin.ParseUint64(buf[offset:]))
		}
	case "uint64":
		if len(buf[offset:]) < 8 {
			return nil, fmt.Errorf("uint64长度不足8:%d", l)
		}
		if p.LittleEndian {
			ret = bin.ParseUint64LittleEndian(buf[offset:])
		} else {
			ret = bin.ParseUint64(buf[offset:])
		}
	case "float", "float32":
		if len(buf[offset:]) < 4 {
			return nil, fmt.Errorf("float32长度不足4:%d", l)
		}
		if p.LittleEndian {
			ret = bin.ParseFloat32LittleEndian(buf[offset:])
		} else {
			ret = bin.ParseFloat32(buf[offset:])
		}
	case "double", "float64":
		if len(buf[offset:]) < 4 {
			return nil, fmt.Errorf("float64长度不足8:%d", l)
		}
		if p.LittleEndian {
			ret = bin.ParseFloat64LittleEndian(buf[offset:])
		} else {
			ret = bin.ParseFloat64(buf[offset:])
		}
	default:
		return nil, fmt.Errorf("不支持的数据类型 %s", p.Type)
	}

	//倍率
	if p.Rate != 0 && p.Rate != 1 {
		ret = cast.ToFloat64(ret) * p.Rate
	}

	//校准
	if p.Correct != 0 {
		ret = cast.ToFloat64(ret) + p.Correct
	}

	//枚举值
	if len(p.Enumerations) > 0 {
		var err error
		ret, err = p.getEnumValue(cast.ToUint(ret))
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func (p *PointWord) Size() int {
	switch p.Type {
	case "bool", "boolean":
		return 1
	case "int8", "uint8":
		return 1
	case "short", "int16":
		return 1
	case "word", "uint16":
		return 1
	case "int32", "int":
		return 2
	case "qword", "uint32", "uint":
		return 2
	case "int64", "uint64":
		return 4
	case "float", "float32":
		return 2
	case "double", "float64":
		return 4
	default:
		return 1
	}
}
