package modbus

import (
	"encoding/binary"
)

// ============ Modbus RTU ============

type RTUFrame struct {
	Addr byte
	FC   byte
	Data []byte
}

func EncodeRTU(addr, fc byte, pdu []byte) []byte {
	frame := make([]byte, 2+len(pdu)+2)
	frame[0] = addr
	frame[1] = fc
	copy(frame[2:], pdu)
	crc := crc16Modbus(frame[:2+len(pdu)])
	frame[2+len(pdu)] = byte(crc & 0xFF)
	frame[2+len(pdu)+1] = byte(crc >> 8)
	return frame
}

// DecodeRTU 解一个 RTU 帧，返回帧 + 剩余数据 + 错误
//
// 长度推断规则：
//   - exception 响应: fc & 0x80 != 0 → 固定 5 字节 (addr + fc + excCode + crc)
//   - 写响应 (FC5/FC6/FC15/FC16): 固定 8 字节 (addr + fc + 4字节echo + crc)
//   - 读请求 (FC1/2/3/4): 固定 8 字节 (addr + fc + 4字节pdu + crc)
//   - 读响应 (FC1/2/3/4): 变长 —— byteCount 在 buf[2]，总长 = 3 + byteCount + 2(crc)
//
// 对于 FC1/2/3/4，由于无法仅从帧头区分请求/响应，使用启发式：
//   - 如果 buf 至少 3 字节，且 buf[2]（疑似 byteCount）在 1~250 之间
//     并且 3+byteCount+2 <= len(buf)，且 CRC 在该长度处有效 → 判定为响应帧
//   - 否则判定为请求帧（固定 8 字节 pdu=4）
func DecodeRTU(buf []byte) (*RTUFrame, []byte, error) {
	if len(buf) < 4 {
		return nil, buf, ErrNeedMore
	}
	addr := buf[0]
	fc := buf[1]

	// exception 响应
	if fc&0x80 != 0 {
		totalLen := 5
		if len(buf) < totalLen {
			return nil, buf, ErrNeedMore
		}
		if !checkRTUCrc(buf[:totalLen]) {
			return nil, buf, ErrCRC
		}
		return &RTUFrame{
			Addr: addr, FC: fc,
			Data: append([]byte(nil), buf[2:3]...),
		}, buf[totalLen:], nil
	}

	// 写单请求/响应 (FC5/FC6): 固定 8 字节
	switch fc {
	case FCWriteSingleCoil, FCWriteSingleRegister:
		totalLen := 2 + 4 + 2
		if len(buf) < totalLen {
			return nil, buf, ErrNeedMore
		}
		if !checkRTUCrc(buf[:totalLen]) {
			return nil, buf, ErrCRC
		}
		return &RTUFrame{
			Addr: addr, FC: fc,
			Data: append([]byte(nil), buf[2:6]...),
		}, buf[totalLen:], nil
	}

	// 写多请求 (FC0F/FC16): 变长 —— 先尝试请求帧，再尝试固定 8 字节响应帧
	if fc == FCWriteMultipleCoils || fc == FCWriteMultipleRegs {
		// 请求帧：addr + fc + startAddr(2) + qty(2) + byteCount(1) + data(N) + crc(2)
		if len(buf) >= 7 {
			byteCount := int(buf[6])
			reqLen := 9 + byteCount
			if byteCount >= 1 && byteCount <= 250 && len(buf) >= reqLen && checkRTUCrc(buf[:reqLen]) {
				return &RTUFrame{
					Addr: addr, FC: fc,
					Data: append([]byte(nil), buf[2:7+byteCount]...),
				}, buf[reqLen:], nil
			}
		}
		// 退回响应帧：固定 8 字节（echo startAddr + qty）
		fixedLen := 8
		if len(buf) < fixedLen {
			return nil, buf, ErrNeedMore
		}
		if !checkRTUCrc(buf[:fixedLen]) {
			return nil, buf, ErrCRC
		}
		return &RTUFrame{
			Addr: addr, FC: fc,
			Data: append([]byte(nil), buf[2:6]...),
		}, buf[fixedLen:], nil
	}

	// 读 FC (FC1/2/3/4): 启发式判定请求 vs 响应
	if fc == FCReadCoils || fc == FCReadDiscreteInputs ||
		fc == FCReadHoldingRegisters || fc == FCReadInputRegisters {

		// 先尝试响应帧格式（byteCount 在 buf[2]）
		if len(buf) >= 3 {
			byteCount := int(buf[2])
			respLen := 3 + byteCount + 2 // addr + fc + byteCount + data + crc
			// byteCount 合法范围（Modbus 规范: 最多 125 寄存器 = 250 bytes）
			if byteCount >= 1 && byteCount <= 250 && len(buf) >= respLen && checkRTUCrc(buf[:respLen]) {
				return &RTUFrame{
					Addr: addr, FC: fc,
					Data: append([]byte(nil), buf[2:2+1+byteCount]...),
				}, buf[respLen:], nil
			}
		}

		// 退回请求帧格式（固定 8 字节）
		reqLen := 2 + 4 + 2
		if len(buf) < reqLen {
			return nil, buf, ErrNeedMore
		}
		if !checkRTUCrc(buf[:reqLen]) {
			return nil, buf, ErrCRC
		}
		return &RTUFrame{
			Addr: addr, FC: fc,
			Data: append([]byte(nil), buf[2:6]...),
		}, buf[reqLen:], nil
	}

	return nil, buf, ErrIllegalFunction
}

// checkRTUCrc 验证 RTU 帧 CRC
func checkRTUCrc(frame []byte) bool {
	if len(frame) < 3 {
		return false
	}
	expected := binary.LittleEndian.Uint16(frame[len(frame)-2 : len(frame)])
	calc := crc16Modbus(frame[:len(frame)-2])
	return expected == calc
}
