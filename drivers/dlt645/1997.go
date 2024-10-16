package dlt645

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var Point1997 = []*Point{
	{"9010", "正向有功总电能", "pos_positive_power", 2, "kWh"},
	{"9020", "反向有功总电能", "neg_positive_power", 2, "kWh"},
	{"B611", "A相电压", "a_voltage", 0, "V"},
	{"B612", "B相电压", "b_voltage", 0, "V"},
	{"B613", "C相电压", "c_voltage", 0, "V"},
	{"B621", "A相电流", "a_current", 2, "A"},
	{"B622", "B相电流", "b_current", 2, "A"},
	{"B623", "C相电流", "c_current", 2, "A"},
	{"B630", "瞬时总功率", "positive_power", 4, "kW"},
	{"B631", "瞬时A相有功功率", "a_positive_power", 4, "kW"},
	{"B632", "瞬时B相有功功率", "b_positive_power", 4, "kW"},
	{"B633", "瞬时C相有功功率", "c_positive_power", 4, "kW"},
	{"B640", "瞬时总无功功率", "reactive_power", 2, "kvar"},
	{"B641", "瞬时A相总无功功率", "a_reactive_power", 2, "kvar"},
	{"B642", "瞬时B相总无功功率", "b_reactive_power", 2, "kvar"},
	{"B643", "瞬时C相总无功功率", "c_reactive_power", 2, "kvar"},
	{"B650", "总功率因数", "influence", 3, ""},
	{"B651", "A相功率因数", "a_influence", 3, ""},
	{"B652", "B相功率因数", "b_influence", 3, ""},
	{"B653", "C相功率因数", "c_influence", 3, ""},
}

func ParsePoint1997(id uint16, buf []byte) (value any, err error) {

	switch (id >> 8) & 0xF0 {
	//电能量数据
	case 0x90:
		return ParseBCD(buf, 4, 2)
	//最大需量数据
	case 0xA0:
		return ParseBCD(buf, 3, 4)
	}

	switch (id >> 8) & 0xFF {

	//最大需量发生时间 MMDDHHmm 月 日 时 分
	case 0xB0, 0xB1, 0xB4, 0xB5, 0xB8, 0xB9:
		return ParseTime(buf)

	//B2 B3 B6 变量数据标识编码表
	case 0xB2:
		switch id & 0xFF {
		case 0x10, 0x11:
			return ParseTime(buf) //MMDDHHmm 编程时间，清零时间
		case 0x12, 0x13:
			return ParseBCD(buf, 2, 0) //编程次数，清零次数
		case 0x14:
			return ParseBCD(buf, 3, 0) //min，电池工作时间
		}
	case 0xB3:
		switch id & 0xF0 {
		case 0x10:
			return ParseBCD(buf, 2, 0) //断相次数
		case 0x20:
			return ParseBCD(buf, 3, 0) //min，断相时间
		case 0x30:
			return ParseTime(buf) //MMDDHHmm，断相发生时间
		case 0x40:
			return ParseTime(buf) //MMDDHHmm，断相结束时间
		}
	case 0xB6:
		switch id & 0xF0 {
		case 0x10:
			return ParseBCD(buf, 2, 0) //电压 a b c
		case 0x20:
			return ParseBCD(buf, 2, 2) //电流 a b c
		case 0x30:
			return ParseBCD(buf, 3, 4) //瞬时功率 t a b c
		case 0x40:
			return ParseBCD(buf, 2, 2) //瞬时无功功率 t a b c
		case 0x50:
			return ParseBCD(buf, 2, 3) //功率因数 t a b c
		}
	}

	return nil, errors.New("不支持的指令")
}

func PointSize1997(id uint16) (int, error) {

	switch (id >> 8) & 0xF0 {
	//电能量数据
	case 0x90:
		return 4, nil
	//最大需量数据
	case 0xA0:
		return 3, nil
	}

	switch (id >> 8) & 0xFF {

	//最大需量发生时间 MMDDHHmm 月 日 时 分
	case 0xB0, 0xB1, 0xB4, 0xB5, 0xB8, 0xB9:
		return 4, nil

	//B2 B3 B6 变量数据标识编码表
	case 0xB2:
		switch id & 0xFF {
		case 0x10, 0x11:
			return 4, nil //MMDDHHmm 编程时间，清零时间
		case 0x12, 0x13:
			return 2, nil //编程次数，清零次数
		case 0x14:
			return 3, nil //min，电池工作时间
		}
	case 0xB3:
		switch id & 0xF0 {
		case 0x10:
			return 2, nil //断相次数
		case 0x20:
			return 3, nil //min，断相时间
		case 0x30:
			return 4, nil //MMDDHHmm，断相发生时间
		case 0x40:
			return 4, nil //MMDDHHmm，断相结束时间
		}
	case 0xB6:
		switch id & 0xF0 {
		case 0x10:
			return 2, nil //电压 a b c
		case 0x20:
			return 2, nil //电流 a b c
		case 0x30:
			return 3, nil //瞬时功率 t a b c
		case 0x40:
			return 2, nil //瞬时无功功率 t a b c
		case 0x50:
			return 2, nil //功率因数 t a b c
		}
	}

	return 0, fmt.Errorf("未支持的指令 %x", id)
}

type dlt1997 struct {
	messenger connect.Messenger
	buf       []byte
}

func NewDlt1997(tunnel connect.Tunnel, opts types.Options) *dlt1997 {
	return &dlt1997{
		messenger: connect.Messenger{
			Timeout: time.Millisecond * time.Duration(opts.Int64("timeout", 1000)),
			Tunnel:  tunnel,
		},
		buf: make([]byte, opts.Int("buffer", 256)),
	}
}

func (m *dlt1997) Read(station string, id string) (any, error) {
	addr, err := ParseAddress(station)
	if err != nil {
		return nil, err
	}
	pack := Pack{
		Address: [6]byte(addr),
		Code:    0x11,
		Data:    nil,
	}

	id2, err := hex.DecodeString(id)
	if err != nil {
		return nil, err
	}
	if len(id2) != 2 {
		return nil, errors.New("标识错误")
	}

	pack.Data = id2

	buf, err := pack.Encode()
	if err != nil {
		return nil, err
	}

	n, err := m.messenger.AskAtLeast(buf, m.buf, 12)
	if err != nil {
		return nil, err
	}
	data := buf[:n]
	err = pack.Decode(data)
	if err != nil {
		return nil, err
	}

	id3 := bin.ParseUint16(id2)

	return ParsePoint1997(id3, pack.Data[2:]) //前2个为寄存器地址
}
