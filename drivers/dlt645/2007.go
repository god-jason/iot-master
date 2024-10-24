package dlt645

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/god-jason/bucket/pkg/bin"
	"github.com/god-jason/bucket/types"
	"github.com/god-jason/iot-master/connect"
	"time"
)

var Point2007 = []*Point{
	{"00000000", "组合有功总电能", "total_power", 2, "kWh"},
	{"00010000", "正向有功总电能", "pos_positive_power", 2, "kWh"},
	{"00020000", "反向有功总电能", "neg_positive_power", 2, "kWh"},
	{"02010100", "A相电压", "a_voltage", 1, "V"},
	{"02010200", "B相电压", "b_voltage", 1, "V"},
	{"02010300", "C相电压", "c_voltage", 1, "V"},
	{"02020100", "A相电流", "a_current", 3, "A"},
	{"02020200", "B相电流", "b_current", 3, "A"},
	{"02020300", "C相电流", "c_current", 3, "A"},
	{"02030000", "瞬时总有功功率", "positive_power", 4, "kW"},
	{"02030100", "瞬时A相有功功率", "a_positive_power", 4, "kW"},
	{"02030200", "瞬时B相有功功率", "b_positive_power", 4, "kW"},
	{"02030300", "瞬时C相有功功率", "c_positive_power", 4, "kW"},
	{"02040000", "瞬时总无功功率", "reactive_power", 4, "kvar"},
	{"02040100", "瞬时A相总无功功率", "a_reactive_power", 4, "kvar"},
	{"02040200", "瞬时B相总无功功率", "b_reactive_power", 4, "kvar"},
	{"02040300", "瞬时C相总无功功率", "c_reactive_power", 4, "kvar"},
	{"02050000", "瞬时总视在功率", "apparent_power", 4, "kVA"},
	{"02050100", "A相视在功率", "a_apparent_power", 4, "kVA"},
	{"02050200", "B相视在功率", "b_apparent_power", 4, "kVA"},
	{"02050300", "C相视在功率", "c_apparent_power", 4, "kVA"},
	{"02060000", "总功率因数", "influence", 3, ""},
	{"02060100", "A相功率因数", "a_influence", 3, ""},
	{"02060200", "B相功率因数", "b_influence", 3, ""},
	{"02060300", "C相功率因数", "c_influence", 3, ""},
}

func ParsePoint2007(id uint32, buf []byte) (value any, err error) {
	switch (id >> 24) & 0xFF {
	case 0x00: //电能量数据标
		return ParseBCD(buf, 4, 2)
	case 0x01: //最大需量 及 时间
		//return ParseBCD(buf, 3, 4)
		return ParseTime2007(buf[3:])
	case 0x02:
		switch (id >> 16) & 0xFF {
		case 0x01:
			return ParseBCD(buf, 2, 1) //电压
		case 0x02:
			return ParseBCD(buf, 3, 3) // 电流
		case 0x03:
			return ParseBCD(buf, 3, 4) //瞬时有功功率
		case 0x04:
			return ParseBCD(buf, 3, 4) //瞬时无功功率
		case 0x05:
			return ParseBCD(buf, 3, 4) //瞬时视在功率
		case 0x06:
			return ParseBCD(buf, 2, 3) //功率因数
		case 0x07:
			return ParseBCD(buf, 2, 1) //相角
		case 0x08, 0x09, 0x0A, 0x0B:
			return ParseBCD(buf, 2, 2)
		case 0x80:
			switch (id >> 8) & 0xFF {
			case 0x00:
				switch id & 0xFF {
				case 0x01:
					return ParseBCD(buf, 3, 3) //零线电流
				case 0x02:
					return ParseBCD(buf, 2, 2) //电网频率
				case 0x03, 0x04, 0x05, 0x06:
					return ParseBCD(buf, 3, 4) //1分钟平均功率 有功需量 无功需量 视在需量
				case 0x07:
					return ParseBCD(buf, 2, 1) //表内温度
				case 0x08, 0x09:
					return ParseBCD(buf, 2, 2) //电池电压
				case 0x0A:
					return ParseBCD(buf, 4, 0) //内部电池工作时间
				case 0x0B:
					return ParseBCD(buf, 4, 4) // 当前阶梯电价
				}
			case 0x01, 0x02, 0x03:
				return ParseBCD(buf, 2, 2)
			}
		}
	case 0x03, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F: //事件记录
	case 0x40: //参变量
	case 0x50: //冻结数据
	case 0x60: //负荷记录
	case 0x70: //安全认证专用数据（抄表）
	}
	return nil, fmt.Errorf("未支持的指令 %0x2f", id)
}

func PointSize2007(id uint32) (int, error) {
	switch (id >> 24) & 0xFF {
	case 0x00: //电能量数据标
		return 4, nil
	case 0x01: //最大需量 及 时间
		return 8, nil
	case 0x02:
		switch (id >> 16) & 0xFF {
		case 0x01:
			return 2, nil //电压
		case 0x02:
			return 3, nil // 电流
		case 0x03:
			return 3, nil //瞬时有功功率
		case 0x04:
			return 3, nil //瞬时无功功率
		case 0x05:
			return 3, nil //瞬时视在功率
		case 0x06:
			return 2, nil //功率因数
		case 0x07:
			return 2, nil //相角
		case 0x08, 0x09, 0x0A, 0x0B:
			return 2, nil
		case 0x80:
			switch (id >> 8) & 0xFF {
			case 0x00:
				switch id & 0xFF {
				case 0x01:
					return 3, nil //零线电流
				case 0x02:
					return 2, nil //电网频率
				case 0x03, 0x04, 0x05, 0x06:
					return 3, nil //1分钟平均功率 有功需量 无功需量 视在需量
				case 0x07:
					return 2, nil //表内温度
				case 0x08, 0x09:
					return 2, nil //电池电压
				case 0x0A:
					return 4, nil //内部电池工作时间
				case 0x0B:
					return 4, nil // 当前阶梯电价
				}
			case 0x01, 0x02, 0x03:
				return 2, nil
			}
		}
	}
	return 0, fmt.Errorf("未支持的指令 %x", id)
}

type dlt2007 struct {
	messenger connect.Messenger
	buf       []byte
}

func NewDlt2007(tunnel connect.Conn, opts types.Options) *dlt2007 {
	return &dlt2007{
		messenger: connect.Messenger{
			Timeout: time.Millisecond * time.Duration(opts.Int64("timeout", 1000)),
			Conn:    tunnel,
		},
		buf: make([]byte, opts.Int("buffer", 256)),
	}
}

func (m *dlt2007) Read(station string, id string) (any, error) {
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
	if len(id2) != 4 {
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

	id3 := bin.ParseUint32(id2)

	return ParsePoint2007(id3, pack.Data[4:]) //前4个为寄存器地址
}
