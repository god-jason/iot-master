// Package iec IEC 60870-5-104 协议
package iec

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

const (
	M_SP_NA_1 byte = 0x01
	M_SP_QA_1 byte = 0x02
	M_ME_NA_1 byte = 0x09
	M_ME_NB_1 byte = 0x0A
	M_ME_NC_1 byte = 0x0B
	C_SC_NA_1 byte = 0x45
	C_DC_NA_1 byte = 0x46
	C_SE_NC_1 byte = 0x4A
)

const (
	COTPeriodic          byte = 0x01
	COTSpontaneous       byte = 0x03
	COTActivation        byte = 0x06
	COTActivationConfirm byte = 0x07
)

const IEC104Start byte = 0x68

type IFrame struct {
	SendSeq uint16
	RecvSeq uint16
	ASDU    []byte
}
type UFrame struct {
	Command byte
}
type SFrame struct {
	RecvSeq uint16
}

func EncodeI(sendSeq, recvSeq uint16, asdu []byte) []byte {
	out := []byte{IEC104Start}
	apduLen := byte(2 + 2 + len(asdu))
	out = append(out, apduLen)
	out = append(out, byte(sendSeq<<1), byte(sendSeq>>7))
	out = append(out, byte(recvSeq<<1), byte(recvSeq>>7))
	out = append(out, asdu...)
	return out
}

func EncodeU(cmd byte) []byte {
	return []byte{IEC104Start, 4, cmd, 0, 0, 0}
}

func DecodeFrame(data []byte) (frameType byte, i IFrame, u UFrame, s SFrame, remaining []byte, err error) {
	if len(data) < 2 {
		err = protocol.ErrNeedMore
		return
	}
	if data[0] != IEC104Start {
		err = fmt.Errorf("iec104: SOH 错误 0x%02X", data[0])
		return
	}
	apduLen := int(data[1])
	totalLen := 2 + apduLen
	if len(data) < totalLen {
		err = protocol.ErrNeedMore
		return
	}
	body := data[2:totalLen]
	remaining = data[totalLen:]

	switch {
	case len(body) == 2:
		frameType = 'S'
		s.RecvSeq = uint16(body[1])<<8 | uint16(body[0]&0xFE)
	case len(body) == 4 && body[0]&0x03 == 0x03:
		frameType = 'U'
		u.Command = body[0]
	case len(body) >= 4:
		frameType = 'I'
		i.SendSeq = uint16(body[1])<<8 | uint16(body[0]>>1)
		i.RecvSeq = uint16(body[3])<<8 | uint16(body[2]>>1)
		if len(body) > 4 {
			i.ASDU = body[4:]
		}
	default:
		err = fmt.Errorf("iec104: 未知帧类型")
	}
	return
}

// ============ Protocol ============

type Protocol struct {
	link     io.ReadWriteCloser
	onValues protocol.OnValues
	mu       sync.Mutex
}

func (*Protocol) Name() string { return protocol.IEC104 }

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
	return nil, fmt.Errorf("iec104 Poll: 尚未实现")
}
func (c *Protocol) Read(deviceId string, key string, timeout time.Duration) (any, error) {
	return nil, fmt.Errorf("iec104 Read: 尚未实现")
}
func (c *Protocol) Write(deviceId string, key string, value any, timeout time.Duration) error {
	return fmt.Errorf("iec104 Write: 尚未实现")
}

func (c *Protocol) OnData(buf []byte) (int, error) {
	// 尚未实现帧切分，消费全部输入
	return len(buf), nil
}

func init() {
	m := protocol.Manifest{
		Name:    "IEC104",
		Version: "60870-5-104",
		Registers: []protocol.Register{
			{Name: "单点遥信 SIQ", Value: "01"},
			{Name: "双点遥信 SPI", Value: "02"},
			{Name: "归一化测量值", Value: "09"},
			{Name: "标度化测量值", Value: "0A"},
			{Name: "短浮点数测量值", Value: "0B"},
			{Name: "单点命令 SC", Value: "45"},
			{Name: "双点命令 DC", Value: "46"},
			{Name: "设点命令 SE", Value: "4A"},
		},
		Formats: []protocol.Format{
			{Name: "SIQ", Value: "siq"},
			{Name: "SPI", Value: "spi"},
			{Name: "归一化值", Value: "normalized"},
			{Name: "标度化值", Value: "scaled"},
			{Name: "float", Value: "float32"},
		},
		Normalizer: func(model *thing.Model) (any, error) {
			return NewMapper(model.Properties)
		},
	}
	m.Factory = func() protocol.Protocol { return &Protocol{} }
	protocol.RegisterProtocol(protocol.IEC104, m)
}

var _ protocol.Protocol = (*Protocol)(nil)
