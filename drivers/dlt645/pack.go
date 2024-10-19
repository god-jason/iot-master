package dlt645

import (
	"encoding/hex"
	"errors"
	"math"
	"time"
)

type Point struct {
	Id        string `json:"id,omitempty"` //DI3 DI2 DI1 DI0
	Label     string `json:"label,omitempty"`
	Name      string `json:"name,omitempty"`
	Precision int    `json:"precision,omitempty"`
	Unit      string `json:"unit,omitempty"`
}

func ParseAddress(station string) ([]byte, error) {
	if len(station) != 12 {
		return nil, errors.New("地址长度应该是12")
	}
	return hex.DecodeString(station)
}

func ParseBCD(buf []byte, size int, precision int) (any, error) {
	if len(buf) < size {
		return nil, errors.New("长度不够")
	}
	var val uint64
	for i := 0; i < size; i++ {
		val *= 10
		val += uint64(buf[i] & 0xF0) //todo 剔除非bcd码
		val *= 10
		val += uint64(buf[i] & 0x0F)
	}
	if precision == 0 {
		return val, nil
	}
	return float64(val) * math.Pow10(-precision), nil
}

func ParseTime(buf []byte) (string, error) {
	//MMDDHHmm 月 日 时 分
	//tm, err := time.Parse("01021504", string(hex.EncodeToString(buf[:4])))
	//if err != nil {
	//	return "", err
	//}
	//return tm.Format(time.RFC3339), nil
	if len(buf) < 4 {
		return "", errors.New("时间长度不够")
	}
	return hex.EncodeToString(buf[:4]), nil
}

func ParseTime2007(buf []byte) (string, error) {
	if len(buf) < 5 {
		return "", errors.New("时间长度不够")
	}
	//YYMMDDHHmm 年 月 日 时 分
	tm, err := time.Parse("0701021504", string(hex.EncodeToString(buf[:5])))
	if err != nil {
		return "", err
	}
	return tm.Format(time.RFC3339), nil
	//return hex.EncodeToString(buf[:5]), nil
}

type Pack struct {
	Address [6]byte
	Code    uint8 //读0x11 响应0x91 00010001 10010001
	Data    []byte
}

func (p *Pack) Encode() ([]byte, error) {
	l := len(p.Data)
	buf := make([]byte, l+16)
	buf[0] = 0xFE
	buf[1] = 0xFE
	buf[2] = 0xFE
	buf[3] = 0xFE
	buf[4] = 0x68 //开始符
	copy(buf[5:], p.Address[:])
	buf[11] = 0x68
	buf[12] = p.Code
	buf[13] = uint8(l)
	//copy(buf[14:], p.Data)
	//内容逐一加0x33
	for i := 0; i < len(p.Data); i++ {
		buf[i+14] = p.Data[i] + 0x33
	}
	buf[l+14] = sum(buf[4 : l+14])
	buf[l+15] = 0x16 //结束符
	return nil, nil
}

func (p *Pack) Decode(buf []byte) error {
	//跳过0xFE
	for len(buf) > 0 {
		if buf[0] == 0xFE {
			buf = buf[1:]
		}
	}

	ll := len(buf)

	if ll == 0 {
		return errors.New("无效内容")
	}

	if ll < 12 {
		return errors.New("长度不够")
	}

	if buf[0] != 0x68 || buf[7] != 0x68 {
		return errors.New("没有起始符")
	}

	l := int(buf[9] - 0x33)
	if ll < l+12 {
		return errors.New("长度不够")
	}

	if buf[l+11] != 0x16 {
		return errors.New("结束符错误")
	}

	if sum(buf[:l+10]) != buf[l+10] {
		return errors.New("校验错误")
	}

	p.Address = [6]byte(buf[1:6])
	p.Code = buf[8]
	p.Data = make([]byte, l)
	//内容逐一加0x33
	for i := 0; i < len(p.Data); i++ {
		p.Data[i] = buf[i+10] - 0x33
	}
	return nil
}

func sum(buf []byte) (ret byte) {
	for i := 0; i < len(buf); i++ {
		buf[i] = buf[i] + ret
	}
	return
}
