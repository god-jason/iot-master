package modbus

import (
	"errors"
	"fmt"
	"github.com/zgwit/iot-master/connect"
	"github.com/zgwit/iot-master/pkg/bin"
	"github.com/zgwit/iot-master/types"
	"time"
)

// TCP Modbus-TCP协议
type TCP struct {
	messenger connect.Messenger
	buf       []byte
}

func NewTCP(conn connect.Conn, opts types.Options) *TCP {
	tcp := &TCP{
		messenger: connect.Messenger{
			Timeout: time.Millisecond * time.Duration(opts.Int64("timeout", 1000)),
			Conn:    conn,
		},
		buf: make([]byte, opts.Int("buffer", 256)),
	}
	return tcp
}

func (m *TCP) execute(cmd []byte) ([]byte, error) {
	bin.WriteUint16(cmd, 0x0A0A) //写入事务ID

	//下发指令
	l, err := m.messenger.AskAtLeast(cmd, m.buf, 10)
	if err != nil {
		return nil, err
	}
	buf := m.buf[:l]

	length := bin.ParseUint16(buf[4:])
	packLen := int(length) + 6
	if packLen > l {
		return nil, errors.New("长度不够")
	}

	//slave := buf[6]
	fc := buf[7]
	//解析错误码
	if fc&0x80 > 0 {
		return nil, fmt.Errorf("错误码：%d", buf[2])
	}

	//解析数据
	//length := 4
	count := int(buf[8])
	switch fc {
	case 1, 2:
		//数组解压
		bb := bin.ExpandBool(buf[9:], count)
		return bb, nil
	case 3, 4, 23:
		return bin.Dup(buf[9:]), nil
	case 5, 15, 6, 16:
		//写指令不处理
		return nil, nil
	default:
		return nil, fmt.Errorf("错误功能码：%d", fc)
	}
}

func (m *TCP) Read(station uint8, code uint8, addr uint16, size uint16) ([]byte, error) {
	b := make([]byte, 12)
	//bin.WriteUint16(b, id)
	bin.WriteUint16(b[2:], 0) //协议版本
	bin.WriteUint16(b[4:], 6) //剩余长度
	b[6] = station
	b[7] = code
	bin.WriteUint16(b[8:], addr)
	bin.WriteUint16(b[10:], uint16(size))

	return m.execute(b)
}

func (m *TCP) Write(station uint8, code uint8, addr uint16, buf []byte) error {
	length := len(buf)
	switch code {
	case 1:
		//如果是线圈，需要Shrink
		if length == 1 {
			code = 5
			//数据 转成 0x0000 0xFF00
			if buf[0] > 0 {
				buf = []byte{0xFF, 0}
			} else {
				buf = []byte{0, 0}
			}
		} else {
			code = 15 //0x0F
			//数组压缩
			b := bin.ShrinkBool(buf)
			count := len(b)
			buf = make([]byte, 3+count)
			bin.WriteUint16(buf, uint16(length))
			buf[2] = uint8(count)
			copy(buf[3:], b)
		}
	case 3:
		if length == 2 {
			code = 6
		} else {
			code = 16 //0x10
			b := make([]byte, 3+length)
			bin.WriteUint16(b, uint16(length/2))
			b[2] = uint8(length)
			copy(b[3:], buf)
			buf = b
		}
	default:
		return errors.New("功能码不支持")
	}

	l := 10 + len(buf)
	b := make([]byte, l)
	//bin.WriteUint16(b, id)
	bin.WriteUint16(b[2:], 0) //协议版本
	bin.WriteUint16(b[4:], 6) //剩余长度
	b[6] = station
	b[7] = code
	bin.WriteUint16(b[8:], addr)
	copy(b[10:], buf)

	_, err := m.execute(b)
	return err
}
