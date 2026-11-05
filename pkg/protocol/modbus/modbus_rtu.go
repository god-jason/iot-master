package modbus

import (
	"github.com/god-jason/iot-master/pkg/protocol"
)

// =====================================================================
// RTUProtocol —— Modbus RTU 实现（串口帧格式 over TCP）
//
// 只负责协议差异部分：按 slave:fc 登记 pending、RTU 帧编码、OnData 解码。
// 请求的串行化由 baseProtocol.serial 统一处理（RTU 半双工，避免 pending 相互覆盖）。
// Attach/Close/Mount/UnMount/Poll/Read/Write/Name 均由 baseProtocol 提供。
// =====================================================================

type RTUProtocol struct {
	baseProtocol
}

func NewRTUProtocol() *RTUProtocol {
	c := &RTUProtocol{}
	// serial=true：同一时刻只允许一个 in-flight 请求
	c.baseProtocol = newBaseProtocol(protocol.RTU, c, true)
	return c
}

// =====================================================================
// transport 实现（RTU 按 slave:fc 匹配 pending）
// =====================================================================

func (c *RTUProtocol) register(slave, fc byte) (reqCtx, chan *Frame, func()) {
	ch, cancel := c.registerPendingRTU(slave, fc)
	return reqCtx{slave: slave}, ch, cancel
}

func (c *RTUProtocol) encode(ctx reqCtx, fc byte, pdu []byte) []byte {
	return EncodeRTU(ctx.slave, fc, pdu)
}

// registerPendingRTU 登记一个 RTU pending 请求（按 slave:fc 匹配）
func (c *RTUProtocol) registerPendingRTU(slave, fc byte) (chan *Frame, func()) {
	return c.registerPendingByKey(rtuPendingKey(slave, fc))
}

// OnData 处理一段字节流，返回消费字节数（跨帧拼包由调用层负责）。
func (c *RTUProtocol) OnData(buf []byte) (int, error) {
	if len(buf) < 4 {
		return 0, nil
	}
	rtu, remaining, err := DecodeRTU(buf)
	switch err {
	case nil:
		frameLen := len(buf) - len(remaining)
		raw := make([]byte, frameLen)
		copy(raw, buf[:frameLen])
		f := BuildFrameFromPdu(protocol.RTU, rtu.Addr, 0, rtu.FC, rtu.Data, raw)

		c.pendingMu.Lock()
		key := rtuPendingKey(rtu.Addr, rtu.FC)
		if req, ok := c.pending[key]; ok {
			select {
			case req.ch <- f:
			default:
			}
			delete(c.pending, key)
		}
		c.pendingMu.Unlock()

		return frameLen, nil
	case ErrNeedMore:
		return 0, nil
	default:
		// CRC 错误或非法帧，跳过 1 字节交由调用层重同步
		return 1, nil
	}
}

func init() {
	m := modbusRTUManifest
	m.Factory = func() protocol.Protocol { return NewRTUProtocol() }
	protocol.RegisterProtocol(protocol.RTU, m)
}

var _ protocol.Protocol = (*RTUProtocol)(nil)
var _ transport = (*RTUProtocol)(nil)
