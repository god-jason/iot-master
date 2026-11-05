// Package cjt188 建设部 CJ/T 188 水表/燃气表协议
//
// 帧结构（对齐 Lua cjt188.lua）：
//
//	[ FE FE FE FE ]  前导码（可选）
//	[ 0x68 ]          SOI
//	[ TYPE 1B ]       仪表类型（0x20=水表）
//	[ ADDR 7B ]       表地址，BCD，低位在前
//	[ CTRL 1B ]       控制字（应答 = 请求 | 0x80）
//	[ LEN 1B ]        长度（DI 2B + SEQ 1B + payload 长度）
//	[ DI 2B ]         数据标识，低位在前
//	[ SEQ 1B ]        序号（自增 0~255）
//	[ payload ]       可选数据
//	[ CHK 1B ]        校验和 = sum(SOI..payload) & 0xFF
//	[ 0x16 ]          EOI
package cjt188

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

// ============================================================
// 常量
// ============================================================

const (
	SOI    byte = 0x68
	EOI    byte = 0x16
	PFE    byte = 0xFE // 前导码
	PFECnt      = 4

	// 仪表类型（常见值）
	TypeWater byte = 0x20 // 水表
	TypeGas   byte = 0x30 // 燃气表
	TypeHeat  byte = 0x33 // 热表
	TypeElec  byte = 0x40 // 电能表

	// 控制字（请求）
	CtrlReadData   byte = 0x01 // 读数据
	CtrlWriteData  byte = 0x02 // 写数据
	CtrlReadTime   byte = 0x03 // 读时间
	CtrlWriteTime  byte = 0x04 // 写时间
	CtrlReadAddr   byte = 0x05 // 读地址
	CtrlWriteValve byte = 0x06 // 写阀门

	// 应答标志：应答 Ctrl = 请求 Ctrl | 0x80
	CtrlReplyMask byte = 0x80
)

// ============================================================
// 帧 codec
// ============================================================

// Frame CJT188 帧
type Frame struct {
	Type byte    // 仪表类型
	Addr [7]byte // 表地址（低位在前存储）
	Ctrl byte    // 控制字
	DI   [2]byte // 数据标识（低位在前存储）
	Seq  byte    // 序号
	Data []byte  // payload（不含 DI/SEQ）
}

// sumChecksum 累加和 & 0xFF
func sumChecksum(b []byte) byte {
	var s byte
	for _, v := range b {
		s += v
	}
	return s
}

// Encode 编码成完整报文（含 FEFEFEFE 前导码）
func (f *Frame) Encode() []byte {
	// 长度字段：DI(2) + SEQ(1) + len(Data)
	length := byte(3 + len(f.Data))

	payload := make([]byte, 0, 1+1+7+1+1+2+1+len(f.Data))
	payload = append(payload, SOI)
	payload = append(payload, f.Type)
	payload = append(payload, f.Addr[:]...)
	payload = append(payload, f.Ctrl, length)
	payload = append(payload, f.DI[0], f.DI[1]) // 低位在前
	payload = append(payload, f.Seq)
	payload = append(payload, f.Data...)

	chk := sumChecksum(payload)

	frame := make([]byte, 0, PFECnt+len(payload)+2)
	for i := 0; i < PFECnt; i++ {
		frame = append(frame, PFE)
	}
	frame = append(frame, payload...)
	frame = append(frame, chk, EOI)
	return frame
}

// Decode 从字节流里切一帧出来
// 对齐 Lua ask() 响应处理
func Decode(data []byte) (*Frame, []byte, error) {
	// 找 SOI（跳过前导 FE/FF）
	soiIdx := -1
	for i := 0; i < len(data); i++ {
		if data[i] == PFE || data[i] == 0xFF {
			continue
		}
		if data[i] == SOI {
			soiIdx = i
			break
		}
		return nil, nil, fmt.Errorf("cjt188: 非法起始字节 0x%02X", data[i])
	}
	if soiIdx == -1 {
		return nil, nil, protocol.ErrNeedMore
	}

	// 帧头最少长度: SOI(1)+TYPE(1)+ADDR(7)+CTRL(1)+LEN(1) = 11
	const headerLen = 11
	if len(data)-soiIdx < headerLen {
		return nil, nil, protocol.ErrNeedMore
	}

	off := soiIdx
	f := &Frame{}
	f.Type = data[off+1]
	copy(f.Addr[:], data[off+2:off+9])
	f.Ctrl = data[off+9]
	length := int(data[off+10])

	// 总长度 = headerLen + DI(2) + SEQ(1) + payload + CHK(1) + EOI(1)
	// length 字段已经包含 DI(2) + SEQ(1) + payload
	totalLen := headerLen + length + 2
	if len(data)-off < totalLen {
		return nil, nil, protocol.ErrNeedMore
	}

	payload := data[off : off+totalLen]
	chkExpected := payload[len(payload)-2]
	chkActual := sumChecksum(payload[:len(payload)-2])
	if chkExpected != chkActual {
		return nil, nil, fmt.Errorf("cjt188: 校验错误 want=0x%02X got=0x%02X", chkExpected, chkActual)
	}
	if payload[len(payload)-1] != EOI {
		return nil, nil, fmt.Errorf("cjt188: EOI 错误 0x%02X", payload[len(payload)-1])
	}

	// 提取 DI（2B）、SEQ（1B）、payload
	// payload 内部: [DI_L, DI_H, SEQ, payload...]
	inner := payload[headerLen : headerLen+length]
	if len(inner) < 3 {
		return nil, nil, fmt.Errorf("cjt188: 内部长度不足")
	}
	f.DI[0] = inner[0]
	f.DI[1] = inner[1]
	f.Seq = inner[2]
	if len(inner) > 3 {
		f.Data = inner[3:]
	}

	return f, data[off+totalLen:], nil
}

// IsReply 判断 Ctrl 是否为应答帧
func (f *Frame) IsReply() bool { return f.Ctrl&CtrlReplyMask != 0 }

// ============================================================
// 协议实例
// ============================================================

// Protocol CJT188 协议实例
type Protocol struct {
	link     io.ReadWriteCloser
	onValues protocol.OnValues
	mu       sync.Mutex

	mounted    sync.Map // deviceId → [7]byte addr
	mountModel sync.Map // deviceId → modelId

	pendingMu sync.Mutex
	pending   map[string]*pendingRequest

	seq     byte // 自增序号
	timeout time.Duration
}

type cjtCachedModel struct {
	mappers []*Cjt188Mapper
}

type pendingRequest struct {
	key string
	ch  chan *Frame
}

func NewProtocol() *Protocol {
	return &Protocol{
		pending: make(map[string]*pendingRequest),
		timeout: 3 * time.Second,
	}
}

func (*Protocol) Name() string { return protocol.CJT188 }

func (c *Protocol) Attach(link io.ReadWriteCloser, onValues protocol.OnValues) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.link = link
	c.onValues = onValues
	return nil
}

func (c *Protocol) Close() error {
	c.pendingMu.Lock()
	for _, p := range c.pending {
		close(p.ch)
	}
	c.pending = make(map[string]*pendingRequest)
	c.pendingMu.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mounted = sync.Map{}
	c.mountModel = sync.Map{}
	c.link = nil
	return nil
}

func (c *Protocol) Mount(deviceId string, address string, modelId string) error {
	addr, err := parse7ByteAddr(address)
	if err != nil {
		return err
	}
	c.mounted.Store(deviceId, addr)
	c.mountModel.Store(deviceId, modelId)
	return nil
}

func (c *Protocol) UnMount(deviceId string) {
	c.mounted.Delete(deviceId)
	c.mountModel.Delete(deviceId)
}

// nextSeq 下一个序号
func (c *Protocol) nextSeq() byte {
	c.seq++
	if c.seq == 0 {
		c.seq = 1
	}
	return c.seq
}

// Poll 轮询所有点位 —— 每个 DI 单独发一次请求（对齐 Lua Device.poll 逐点 read）
func (c *Protocol) Poll(deviceId string, timeout time.Duration) (map[string]any, error) {
	to := c.timeout
	if timeout > 0 {
		to = timeout
	}

	addrAny, ok := c.mounted.Load(deviceId)
	if !ok {
		return nil, protocol.ErrNotMounted
	}
	addr := addrAny.([7]byte)

	modelIdAny, _ := c.mountModel.Load(deviceId)
	cached, ok := c.findCached(modelIdAny.(string))
	if !ok {
		return nil, fmt.Errorf("cjt188: 设备 %s 未注册物模型", deviceId)
	}

	results := make(map[string]any)
	for _, mapper := range cached.mappers {
		for _, req := range mapper.EncodeRequests() {
			// req: { Ctrl, DI, payload }
			key := pendingKey(addr, req.Ctrl, req.DI)

			c.mu.Lock()
			seq := c.nextSeq()
			c.mu.Unlock()

			frame := &Frame{
				Type: TypeWater,
				Addr: addr,
				Ctrl: req.Ctrl,
				DI:   req.DI,
				Seq:  seq,
				Data: req.Payload,
			}

			c.pendingMu.Lock()
			ch := make(chan *Frame, 1)
			c.pending[key] = &pendingRequest{key: key, ch: ch}
			c.pendingMu.Unlock()

			c.mu.Lock()
			var err error
			if c.link != nil {
				_, err = c.link.Write(frame.Encode())
			} else {
				err = fmt.Errorf("cjt188: 未 Attach 连接")
			}
			c.mu.Unlock()

			if err != nil {
				c.pendingMu.Lock()
				delete(c.pending, key)
				c.pendingMu.Unlock()
				return nil, err
			}

			select {
			case resp := <-ch:
				c.pendingMu.Lock()
				delete(c.pending, key)
				c.pendingMu.Unlock()
				if resp == nil {
					return nil, protocol.ErrTimeout
				}
				// 验证应答 Ctrl = request Ctrl | 0x80
				expectedCtrl := req.Ctrl | CtrlReplyMask
				if resp.Ctrl != expectedCtrl {
					return nil, fmt.Errorf("cjt188: 非正常应答 want ctrl=0x%02X got=0x%02X", expectedCtrl, resp.Ctrl)
				}
				vals, err := mapper.DecodeResponse(req.DI, resp.Data)
				if err != nil {
					return nil, err
				}
				for k, v := range vals {
					results[k] = v
				}
			case <-time.After(to):
				c.pendingMu.Lock()
				delete(c.pending, key)
				c.pendingMu.Unlock()
				return nil, protocol.ErrTimeout
			}
		}
	}
	return results, nil
}

func (c *Protocol) Read(deviceId string, key string, timeout time.Duration) (any, error) {
	vals, err := c.Poll(deviceId, timeout)
	if err != nil {
		return nil, err
	}
	if v, ok := vals[key]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("cjt188: key %q 不存在", key)
}

func (c *Protocol) Write(deviceId string, key string, value any, timeout time.Duration) error {
	addrAny, ok := c.mounted.Load(deviceId)
	if !ok {
		return protocol.ErrNotMounted
	}
	addr := addrAny.([7]byte)

	modelIdAny, _ := c.mountModel.Load(deviceId)
	cached, ok := c.findCached(modelIdAny.(string))
	if !ok {
		return fmt.Errorf("cjt188: 设备 %s 未注册物模型", deviceId)
	}

	// 从 mapper 里找 target item
	var target *MapItem
	for _, m := range cached.mappers {
		if it := m.FindItemByName(key); it != nil {
			target = it
			break
		}
	}
	if target == nil {
		// fallback：默认阀门控制
		target = &MapItem{Ctrl: CtrlWriteValve, DI: [2]byte{0x01, 0x00}}
	}

	var v byte
	switch t := value.(type) {
	case bool:
		if t {
			v = 0x01
		}
	case int:
		v = byte(t)
	case uint8:
		v = t
	default:
		return fmt.Errorf("cjt188 Write: value 类型不支持 %T", value)
	}

	req := Cjt188Request{
		Ctrl:    target.Ctrl,
		DI:      target.DI,
		Payload: []byte{v},
	}

	pkey := pendingKey(addr, req.Ctrl, req.DI)

	c.mu.Lock()
	seq := c.nextSeq()
	c.mu.Unlock()

	frame := &Frame{
		Type: TypeWater,
		Addr: addr,
		Ctrl: req.Ctrl,
		DI:   req.DI,
		Seq:  seq,
		Data: req.Payload,
	}

	c.pendingMu.Lock()
	ch := make(chan *Frame, 1)
	c.pending[pkey] = &pendingRequest{key: pkey, ch: ch}
	c.pendingMu.Unlock()

	c.mu.Lock()
	var err error
	if c.link != nil {
		_, err = c.link.Write(frame.Encode())
	} else {
		err = fmt.Errorf("cjt188: 未 Attach 连接")
	}
	c.mu.Unlock()

	if err != nil {
		c.pendingMu.Lock()
		delete(c.pending, pkey)
		c.pendingMu.Unlock()
		return err
	}

	to := c.timeout
	if timeout > 0 {
		to = timeout
	}
	select {
	case resp := <-ch:
		c.pendingMu.Lock()
		delete(c.pending, pkey)
		c.pendingMu.Unlock()
		if resp == nil {
			return protocol.ErrTimeout
		}
		expectedCtrl := req.Ctrl | CtrlReplyMask
		if resp.Ctrl != expectedCtrl {
			return fmt.Errorf("cjt188 Write: 非正常应答 want ctrl=0x%02X got=0x%02X", expectedCtrl, resp.Ctrl)
		}
		return nil
	case <-time.After(to):
		c.pendingMu.Lock()
		delete(c.pending, pkey)
		c.pendingMu.Unlock()
		return protocol.ErrTimeout
	}
}

// OnData 处理一段字节流，返回消费字节数（跨帧拼包由调用层负责）。
func (c *Protocol) OnData(buf []byte) (int, error) {
	frame, remaining, err := Decode(buf)
	switch err {
	case nil:
		frameLen := len(buf) - len(remaining)

		c.pendingMu.Lock()
		key := pendingKey(frame.Addr, frame.Ctrl&^CtrlReplyMask, frame.DI)
		if p, ok := c.pending[key]; ok {
			select {
			case p.ch <- frame:
			default:
			}
			delete(c.pending, key)
		}
		c.pendingMu.Unlock()
		return frameLen, nil

	case protocol.ErrNeedMore:
		return 0, nil
	default:
		// 非法起始字节等错误，跳过 1 字节交由调用层重试
		return 1, nil
	}
}

func pendingKey(addr [7]byte, ctrl byte, di [2]byte) string {
	return fmt.Sprintf("%x:%02X:%02X%02X", addr[:], ctrl, di[0], di[1])
}

func (c *Protocol) findCached(modelId string) (*cjtCachedModel, bool) {
	v, ok := protocol.GetModel(modelId)
	if !ok {
		return nil, false
	}
	m, ok := v.(*cjtCachedModel)
	return m, ok
}

// ============================================================
// 地址解析
// ============================================================

// parse7ByteAddr 解析表地址 → [7]byte（低位在前存储）
func parse7ByteAddr(s string) ([7]byte, error) {
	var out [7]byte
	s = strings.TrimSpace(s)
	if s == "" {
		return out, fmt.Errorf("地址为空")
	}
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if len(parts) > 7 {
			return out, fmt.Errorf("地址段数 > 7")
		}
		tmp := make([]byte, len(parts))
		for i, p := range parts {
			v, err := parseUint8(strings.TrimSpace(p))
			if err != nil {
				return out, fmt.Errorf("地址段 %q: %w", p, err)
			}
			tmp[i] = v
		}
		// 逆序 → 低位在前存储
		for i := range tmp {
			out[6-i] = tmp[i]
		}
		return out, nil
	}
	var v uint64
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		if _, err = fmt.Sscanf(s, "%x", &v); err != nil {
			return out, fmt.Errorf("地址 %q 无法解析", s)
		}
	}
	for i := 6; i >= 0 && v > 0; i-- {
		out[i] = byte(v & 0xFF)
		v >>= 8
	}
	return out, nil
}

func parseUint8(s string) (byte, error) {
	var v uint64
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		if _, err = fmt.Sscanf(s, "%x", &v); err != nil {
			return 0, err
		}
	}
	if v > 0xFF {
		return 0, fmt.Errorf("超出 uint8 范围")
	}
	return byte(v), nil
}

func init() {
	m := protocol.Manifest{
		Name:    "CJT188",
		Version: "1.0",
		Registers: []protocol.Register{
			{Name: "数据标识符 DI", Value: "DI"},
		},
		Formats: []protocol.Format{
			{Name: "BCD", Value: "bcd"},
			{Name: "uint8", Value: "uint8"},
			{Name: "uint16", Value: "uint16"},
			{Name: "float32", Value: "float32"},
			{Name: "string", Value: "string"},
			{Name: "bool", Value: "bool"},
		},
		Normalizer: func(model *thing.Model) (any, error) {
			mapper, err := NewMapper(model.Properties)
			if err != nil {
				return nil, err
			}
			return &cjtCachedModel{mappers: []*Cjt188Mapper{mapper}}, nil
		},
	}
	m.Factory = func() protocol.Protocol { return NewProtocol() }
	protocol.RegisterProtocol(protocol.CJT188, m)
}

var _ protocol.Protocol = (*Protocol)(nil)
