// Package jt808 部标 JT/T 808-2019 车载终端协议
package jt808

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

const (
	SOI byte = 0x7E
)

type Frame struct {
	MsgID uint16
	Props uint16
	Phone [6]byte
	Seq   uint16
	Body  []byte
	Extra []byte
}

func (f *Frame) Encode() []byte {
	out := []byte{SOI}
	out = append(out, byte(f.MsgID>>8), byte(f.MsgID))
	out = append(out, byte(f.Props>>8), byte(f.Props))
	out = append(out, f.Phone[:]...)
	out = append(out, byte(f.Seq>>8), byte(f.Seq))
	out = append(out, f.Extra...)
	out = append(out, f.Body...)

	var chk byte
	for _, b := range out[1:] {
		chk ^= b
	}
	out = append(out, chk, SOI)

	result := []byte{out[0]}
	for _, b := range out[1 : len(out)-1] {
		switch b {
		case 0x7E:
			result = append(result, 0x7D, 0x01)
		case 0x7D:
			result = append(result, 0x7D, 0x02)
		default:
			result = append(result, b)
		}
	}
	result = append(result, out[len(out)-1])
	return result
}

func Decode(data []byte) (*Frame, []byte, error) {
	if len(data) < 14 {
		return nil, data, protocol.ErrNeedMore
	}
	if data[0] != SOI {
		return nil, data, fmt.Errorf("jt808: SOI 错误")
	}
	eoiIdx := -1
	for i := 1; i < len(data); i++ {
		if data[i] == SOI {
			eoiIdx = i
			break
		}
	}
	if eoiIdx < 0 {
		return nil, data, protocol.ErrNeedMore
	}

	raw := data[1:eoiIdx]
	var payload []byte
	for i := 0; i < len(raw); i++ {
		if raw[i] == 0x7D && i+1 < len(raw) {
			switch raw[i+1] {
			case 0x01:
				payload = append(payload, 0x7E)
			case 0x02:
				payload = append(payload, 0x7D)
			default:
				payload = append(payload, raw[i+1])
			}
			i++
		} else {
			payload = append(payload, raw[i])
		}
	}

	if len(payload) < 11 {
		return nil, data, protocol.ErrNeedMore
	}
	var chk byte
	for _, b := range payload[:len(payload)-1] {
		chk ^= b
	}
	if chk != payload[len(payload)-1] {
		return nil, data, fmt.Errorf("jt808: 校验错误")
	}

	f := &Frame{
		MsgID: uint16(payload[0])<<8 | uint16(payload[1]),
		Props: uint16(payload[2])<<8 | uint16(payload[3]),
		Seq:   uint16(payload[10])<<8 | uint16(payload[11]),
	}
	copy(f.Phone[:], payload[4:10])

	bodyOffset := 12
	if f.Props&0x4000 != 0 {
		if len(payload) < 17 {
			return nil, data, protocol.ErrNeedMore
		}
		f.Extra = payload[12:17]
		bodyOffset = 17
	}
	if bodyOffset < len(payload)-1 {
		f.Body = payload[bodyOffset : len(payload)-1]
	}

	return f, data[eoiIdx+1:], nil
}

// ============ Protocol ============

type Protocol struct {
	link     io.ReadWriteCloser
	onValues protocol.OnValues
	mu       sync.Mutex
}

func (*Protocol) Name() string { return protocol.JT808 }

func (c *Protocol) Attach(link io.ReadWriteCloser, onValues protocol.OnValues) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.link = link
	c.onValues = onValues
	return nil
}
func (c *Protocol) Close() error { c.mu.Lock(); defer c.mu.Unlock(); c.link = nil; return nil }

func (c *Protocol) Mount(deviceId string, address string, modelId string) error { return nil }
func (c *Protocol) UnMount(deviceId string)                                     {}

func (c *Protocol) Poll(deviceId string, timeout time.Duration) (map[string]any, error) {
	return nil, fmt.Errorf("jt808: 不支持主动轮询，请等待终端上报")
}
func (c *Protocol) Read(deviceId string, key string, timeout time.Duration) (any, error) {
	return nil, fmt.Errorf("jt808 Read: 尚未实现")
}
func (c *Protocol) Write(deviceId string, key string, value any, timeout time.Duration) error {
	return fmt.Errorf("jt808 Write: 尚未实现")
}

func (c *Protocol) OnData(buf []byte) (int, error) {
	// 尚未实现帧切分，消费全部输入
	return len(buf), nil
}

func init() {
	m := protocol.Manifest{
		Name:    "JT808",
		Version: "2019",
		Registers: []protocol.Register{
			{Name: "消息ID msgID", Value: "msgID"},
		},
		Formats: []protocol.Format{
			{Name: "二进制", Value: "binary"},
			{Name: "string", Value: "string"},
			{Name: "uint32", Value: "uint32"},
		},
		Normalizer: func(model *thing.Model) (any, error) {
			return NewMapper(model.Properties)
		},
	}
	m.Factory = func() protocol.Protocol { return &Protocol{} }
	protocol.RegisterProtocol(protocol.JT808, m)
}

var _ protocol.Protocol = (*Protocol)(nil)
