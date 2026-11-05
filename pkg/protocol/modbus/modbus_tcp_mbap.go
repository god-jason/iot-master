package modbus

import (
	"encoding/binary"
	"fmt"
)

// ============ Modbus TCP MBAP ============
//
// 本包统一约定：PDU 不含 FC（FC 由帧头或单独参数携带），
// EncodeMBAP / DecodeMBAP / EncodeRTU / DecodeRTU 四者保持一致。

// mbapMaxLength MBAP Length 字段上限。
// Length = UnitID(1) + PDU，而 PDU 最大 253 字节（ADU 上限 260 - MBAP 头 7）→ Length 上限 254。
// 超出该范围的 Length 说明字段已损坏：继续等待更多字节只会让流永久卡死，
// 因此按无效帧处理，交由调用层跳过重同步。
const mbapMaxLength = 254

type MBAPHeader struct {
	TransactionID uint16
	ProtocolID    uint16
	Length        uint16
	UnitID        byte
}

// EncodeMBAP 编码一个 Modbus TCP 帧：MBAP 头(7) + FC(1) + PDU。
// pdu 不含 FC，与 DecodeMBAP 的返回、以及 EncodeRTU 的参数语义一致。
func EncodeMBAP(txID uint16, unitID, fc byte, pdu []byte) []byte {
	length := uint16(2 + len(pdu)) // unitID(1) + fc(1) + pdu
	frame := make([]byte, 8+len(pdu))
	binary.BigEndian.PutUint16(frame[0:2], txID)
	binary.BigEndian.PutUint16(frame[2:4], 0)
	binary.BigEndian.PutUint16(frame[4:6], length)
	frame[6] = unitID
	frame[7] = fc
	copy(frame[8:], pdu)
	return frame
}

// DecodeMBAP 解一个 MBAP 帧，返回 头 + FC + PDU(不含 FC) + 剩余数据 + 错误。
// 数据不完整返回 ErrNeedMore；字段非法（protocol id / length 越界）返回具体错误，
// 便于调用层区分「继续等待」与「需要重同步」。
func DecodeMBAP(buf []byte) (*MBAPHeader, byte, []byte, []byte, error) {
	if len(buf) < 7 {
		return nil, 0, nil, nil, ErrNeedMore
	}
	h := &MBAPHeader{
		TransactionID: binary.BigEndian.Uint16(buf[0:2]),
		ProtocolID:    binary.BigEndian.Uint16(buf[2:4]),
		Length:        binary.BigEndian.Uint16(buf[4:6]),
		UnitID:        buf[6],
	}
	if h.ProtocolID != 0 {
		return h, 0, nil, nil, fmt.Errorf("modbus tcp: invalid protocol id %d", h.ProtocolID)
	}
	if h.Length < 2 || h.Length > mbapMaxLength {
		return h, 0, nil, nil, fmt.Errorf("modbus tcp: invalid length %d (期望 2~%d)", h.Length, mbapMaxLength)
	}
	expectedPDULen := int(h.Length) - 1
	if len(buf[7:]) < expectedPDULen {
		return nil, 0, nil, nil, ErrNeedMore
	}
	pdu := buf[7 : 7+expectedPDULen]
	return h, pdu[0], pdu[1:], buf[7+expectedPDULen:], nil
}
