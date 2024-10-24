package modbus

import (
	"errors"
	"fmt"
	"github.com/god-jason/bucket/pkg/bin"
	"github.com/spf13/cast"
)

type Point interface {
	Encode(data any) ([]byte, error)
	Parse(address uint16, buf []byte) (any, error)
}

type PointBit struct {
	Name    string `json:"name"`    //名称
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
	if offset > l*8 {
		return nil, errors.New("长度不够")
	}

	cur := offset / 8
	bit := offset % 8

	ret := buf[cur] & (1 << bit)

	return ret > 0, nil
}

type PointWord struct {
	Name      string  `json:"name"`              //名称
	Type      string  `json:"type"`              //类型
	Address   uint16  `json:"address"`           //偏移
	BigEndian bool    `json:"be,omitempty"`      //大端模式
	Rate      float64 `json:"rate,omitempty"`    //倍率
	Correct   float64 `json:"correct,omitempty"` //纠正
	Bits      []*Bit  `json:"bits,omitempty"`    //位，1 2 3...
}

type Bit struct {
	Name string `json:"name"` //名称
	Bit  int    `json:"bit"`  //偏移
}

func (p *PointWord) Encode(data any) ([]byte, error) {
	var ret []byte

	//纠正
	if p.Correct != 0 {
		data = cast.ToFloat64(data) - p.Correct
	}

	//倍率逆转换
	if p.Rate != 0 && p.Rate != 1 {
		data = cast.ToFloat64(data) / p.Rate
	}

	switch p.Type {
	case "short", "int16":
		ret = make([]byte, 2)
		val := cast.ToInt16(data)
		if p.BigEndian {
			bin.WriteUint16(ret, uint16(val))
		} else {
			bin.WriteUint16LittleEndian(ret, uint16(val))
		}
	case "word", "uint16":
		ret = make([]byte, 2)
		val := cast.ToUint16(data)
		if p.BigEndian {
			bin.WriteUint16(ret, val)
		} else {
			bin.WriteUint16LittleEndian(ret, val)
		}
	case "int32", "int":
		ret = make([]byte, 4)
		val := cast.ToInt32(data)
		if p.BigEndian {
			bin.WriteUint32(ret, uint32(val))
		} else {
			bin.WriteUint32LittleEndian(ret, uint32(val))
		}
	case "qword", "uint32", "uint":
		ret = make([]byte, 4)
		val := cast.ToUint32(data)
		if p.BigEndian {
			bin.WriteUint32(ret, val)
		} else {
			bin.WriteUint32LittleEndian(ret, val)
		}
	case "float", "float32":
		ret = make([]byte, 4)
		val := cast.ToFloat32(data)
		if p.BigEndian {
			bin.WriteFloat32(ret, val)
		} else {
			bin.WriteFloat32LittleEndian(ret, val)
		}
	case "double", "float64":
		ret = make([]byte, 8)
		val := cast.ToFloat64(data)
		if p.BigEndian {
			bin.WriteFloat64(ret, val)
		} else {
			bin.WriteFloat64LittleEndian(ret, val)
		}
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
	case "short", "int16":
		if p.BigEndian {
			ret = int16(bin.ParseUint16(buf[offset:]))
		} else {
			ret = int16(bin.ParseUint16LittleEndian(buf[offset:]))
		}
		if p.Rate != 0 && p.Rate != 1 {
			ret = float64(ret.(int16)) * p.Rate
		}
		if p.Correct != 0 {
			ret = float64(ret.(int16)) + p.Correct
		}
	case "word", "uint16":
		if p.BigEndian {
			ret = bin.ParseUint16(buf[offset:])
		} else {
			ret = bin.ParseUint16LittleEndian(buf[offset:])
		}
		//取位
		if p.Bits != nil && len(p.Bits) > 0 {
			rets := make(map[string]bool)
			for _, b := range p.Bits {
				rets[b.Name] = (ret.(uint16))&(1<<b.Bit) > 0
			}
			return rets, nil
		}

		if p.Rate != 0 && p.Rate != 1 {
			ret = float64(ret.(uint16)) * p.Rate
		}
		if p.Correct != 0 {
			ret = float64(ret.(int16)) + p.Correct
		}
	case "int32", "int":
		if p.BigEndian {
			ret = int32(bin.ParseUint32(buf[offset:]))
		} else {
			ret = int32(bin.ParseUint32LittleEndian(buf[offset:]))
		}
		if p.Rate != 0 && p.Rate != 1 {
			ret = float64(ret.(int32)) * p.Rate
		}
		if p.Correct != 0 {
			ret = float64(ret.(int16)) + p.Correct
		}
	case "qword", "uint32", "uint":
		if p.BigEndian {
			ret = bin.ParseUint32(buf[offset:])
		} else {
			ret = bin.ParseUint32LittleEndian(buf[offset:])
		}
		if p.Rate != 0 && p.Rate != 1 {
			ret = float64(ret.(uint32)) * p.Rate
		}
		if p.Correct != 0 {
			ret = float64(ret.(int16)) + p.Correct
		}
	case "float", "float32":
		if p.BigEndian {
			ret = bin.ParseFloat32(buf[offset:])
		} else {
			ret = bin.ParseFloat32LittleEndian(buf[offset:])
		}
		if p.Rate != 0 && p.Rate != 1 {
			ret = float64(ret.(float32)) * p.Rate
		}
		if p.Correct != 0 {
			ret = float64(ret.(int16)) + p.Correct
		}
	case "double", "float64":
		if p.BigEndian {
			ret = bin.ParseFloat64(buf[offset:])
		} else {
			ret = bin.ParseFloat64LittleEndian(buf[offset:])
		}
		if p.Rate != 0 && p.Rate != 1 {
			ret = ret.(float64) * p.Rate
		}
		if p.Correct != 0 {
			ret = float64(ret.(int16)) + p.Correct
		}
	default:
		return nil, fmt.Errorf("不支持的数据类型 %s", p.Type)
	}

	return ret, nil
}
