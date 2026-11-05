package modbus

import (
	"fmt"

	"github.com/god-jason/iot-master/pkg/protocol"
)

// =====================================================================
// TCPProtocol —— Modbus TCP 完整实现（MBAP）
//
// 只负责协议差异部分：按 txID 登记 pending、MBAP 帧编码、OnData 解码。
// Attach/Close/Mount/UnMount/Poll/Read/Write/Name 均由 baseProtocol 提供。
// =====================================================================

type TCPProtocol struct {
	baseProtocol
}

func NewTCPProtocol() *TCPProtocol {
	c := &TCPProtocol{}
	// tr 指向自身：register / encode 由本类型实现
	c.baseProtocol = newBaseProtocol(protocol.TCP, c, false)
	return c
}

// =====================================================================
// transport 实现（TCP 按 txID 匹配 pending）
// =====================================================================

// register 登记一个 TCP pending 请求。
// TCP 语义下 slave 参数即 UnitID，回填进 reqCtx.unitID。
func (c *TCPProtocol) register(slave, _ byte) (reqCtx, chan *Frame, func()) {
	txID, ch, cancel := c.registerPending()
	return reqCtx{txID: txID, unitID: slave, slave: slave}, ch, cancel
}

func (c *TCPProtocol) encode(ctx reqCtx, fc byte, pdu []byte) []byte {
	return EncodeMBAP(ctx.txID, ctx.unitID, fc, pdu)
}

// registerPending 登记一个 TCP pending 请求（按 txID 匹配）
func (c *TCPProtocol) registerPending() (uint16, chan *Frame, func()) {
	c.pendingMu.Lock()
	c.txID++
	if c.txID == 0 {
		c.txID = 1
	}
	txID := c.txID
	key := fmt.Sprintf("tcp:%d", txID)
	ch := make(chan *Frame, 1)
	c.pending[key] = &pendingRequest{ch: ch}
	c.pendingMu.Unlock()

	cancel := func() {
		c.pendingMu.Lock()
		delete(c.pending, key)
		c.pendingMu.Unlock()
	}
	return txID, ch, cancel
}

// OnData 处理一段字节流，返回消费字节数（跨帧拼包由调用层负责）。
func (c *TCPProtocol) OnData(buf []byte) (int, error) {
	if len(buf) < 7 {
		return 0, nil
	}
	h, fc, pdu, _, err := DecodeMBAP(buf)
	switch err {
	case nil:
		frameLen := 7 + 1 + len(pdu)
		if len(buf) < frameLen {
			return 0, nil
		}
		raw := make([]byte, frameLen)
		copy(raw, buf[:frameLen])
		f := BuildFrameFromPdu(protocol.TCP, 0, h.UnitID, fc, pdu, raw)

		c.pendingMu.Lock()
		key := fmt.Sprintf("tcp:%d", h.TransactionID)
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
		// 非法帧（protocol id / length 越界等），跳过 1 字节交由调用层重同步
		return 1, nil
	}
}

func init() {
	m := modbusTCPManifest
	m.Factory = func() protocol.Protocol { return NewTCPProtocol() }
	protocol.RegisterProtocol(protocol.TCP, m)
}

var _ protocol.Protocol = (*TCPProtocol)(nil)
var _ transport = (*TCPProtocol)(nil)
