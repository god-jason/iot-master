package cjt188

import (
	"bytes"
	"testing"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/protocol/meter"
	"github.com/god-jason/iot-master/pkg/thing"
)

// ========== Codec round-trip ==========

func TestCodecRoundTrip_Basic(t *testing.T) {
	// 请求帧：FEFEFEFE | 0x68 | type=0x20 | addr(7B 低位在前) | ctrl=0x01 | len | di(2B) | seq | chk | 0x16
	frame := &Frame{
		Type: TypeWater,
		Addr: [7]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}, // 低位在前存储
		Ctrl: CtrlReadData,
		DI:   [2]byte{0x01, 0x00}, // 低位在前: 0x0001
		Seq:  0x42,
		Data: nil,
	}

	encoded := frame.Encode()

	// 前导 FEFEFEFE
	for i := 0; i < PFECnt; i++ {
		if encoded[i] != PFE {
			t.Fatalf("前导[%d] 错误: 0x%02X", i, encoded[i])
		}
	}

	// SOI = 0x68
	if encoded[PFECnt] != SOI {
		t.Fatalf("SOI 错误 0x%02X", encoded[PFECnt])
	}

	// Decode
	decoded, remaining, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode 失败: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("剩余非空: %x", remaining)
	}

	if decoded.Type != frame.Type {
		t.Errorf("Type 错误: want 0x%02X got 0x%02X", frame.Type, decoded.Type)
	}
	if decoded.Ctrl != frame.Ctrl {
		t.Errorf("Ctrl 错误")
	}
	if decoded.Seq != frame.Seq {
		t.Errorf("Seq 错误: want %d got %d", frame.Seq, decoded.Seq)
	}
	if decoded.DI != frame.DI {
		t.Errorf("DI 错误: want %x got %x", frame.DI, decoded.DI)
	}
	if decoded.Addr != frame.Addr {
		t.Errorf("Addr 错误: want %x got %x", frame.Addr, decoded.Addr)
	}
}

func TestCodecRoundTrip_Payload(t *testing.T) {
	// 带 payload
	frame := &Frame{
		Type: TypeWater,
		Addr: [7]byte{0, 0, 0, 0, 0, 0, 1},
		Ctrl: CtrlWriteValve,
		DI:   [2]byte{0x01, 0x00},
		Seq:  1,
		Data: []byte{0x01}, // 阀门开
	}

	encoded := frame.Encode()
	decoded, _, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded.Data, []byte{0x01}) {
		t.Errorf("payload 错误: got %x", decoded.Data)
	}
	if decoded.Ctrl != CtrlWriteValve {
		t.Errorf("Ctrl 错误")
	}
}

func TestCodecRoundTrip_NoPreamble(t *testing.T) {
	// 响应经常不带前导
	frame := &Frame{
		Type: TypeWater,
		Addr: [7]byte{1, 2, 3, 4, 5, 6, 7},
		Ctrl: CtrlReadData | CtrlReplyMask, // 正常应答
		DI:   [2]byte{0x01, 0x00},
		Seq:  0,
		Data: []byte{0x00, 0x12, 0x34, 0x56},
	}
	full := frame.Encode()
	noFE := full[PFECnt:]

	decoded, _, err := Decode(noFE)
	if err != nil {
		t.Fatalf("无前导码 Decode 失败: %v", err)
	}
	if !decoded.IsReply() {
		t.Error("应该是应答帧")
	}
	if decoded.Ctrl != CtrlReadData|CtrlReplyMask {
		t.Errorf("应答 Ctrl 错误")
	}
	if !bytes.Equal(decoded.Data, []byte{0x00, 0x12, 0x34, 0x56}) {
		t.Errorf("Data 错误")
	}
}

func TestCodecDecode_Incomplete(t *testing.T) {
	frame := &Frame{
		Type: TypeWater,
		Addr: [7]byte{1, 2, 3, 4, 5, 6, 7},
		Ctrl: CtrlReadData,
		DI:   [2]byte{0x01, 0x00},
		Seq:  1,
	}
	encoded := frame.Encode()
	_, _, err := Decode(encoded[:len(encoded)-1])
	if err == nil {
		t.Error("截断帧应报错")
	}
}

func TestCodecDecode_BadChecksum(t *testing.T) {
	frame := &Frame{
		Type: TypeWater,
		Addr: [7]byte{1, 2, 3, 4, 5, 6, 7},
		Ctrl: CtrlReadData,
		DI:   [2]byte{0x01, 0x00},
		Seq:  1,
	}
	encoded := frame.Encode()
	encoded[len(encoded)-2] ^= 0xFF
	_, _, err := Decode(encoded)
	if err == nil {
		t.Error("校验错帧应报错")
	}
}

// ========== Mapper ==========

func TestCjtMapper_EncodeDecode(t *testing.T) {
	props := []thing.Property{
		{Key: "volume", Address: "00.01", Format: "XXXXXX.XX", Rate: 1},
	}

	mapper, err := NewMapper(props)
	if err != nil {
		t.Fatal(err)
	}

	reqs := mapper.EncodeRequests()
	if len(reqs) != 1 {
		t.Fatalf("期望 1 个请求 got %d", len(reqs))
	}

	// DI "00.01" → hi=0x00, lo=0x01 → [lo, hi] = [0x01, 0x00]
	expectedDI := [2]byte{0x01, 0x00}
	if reqs[0].DI != expectedDI {
		t.Errorf("DI 错误: want %x got %x", expectedDI, reqs[0].DI)
	}

	// 模拟响应 payload: XXXXXX.XX BCD 值
	respData := []byte{0x00, 0x12, 0x34, 0x56} // BCD 00123456 → 1234.56
	vals, err := mapper.DecodeResponse(expectedDI, respData)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := vals["volume"]; !ok {
		t.Error("缺 volume")
	} else if v.(float64) != 1234.56 {
		t.Errorf("volume 解码错误: got %v", v)
	}
}

func TestParseDI_Cjt(t *testing.T) {
	cases := []struct {
		addr string
		want [2]byte
	}{
		{"00.01", [2]byte{0x01, 0x00}},
		{"01", [2]byte{0x01, 0x00}},
		{"0x0001", [2]byte{0x01, 0x00}},
		{"0x2030", [2]byte{0x30, 0x20}},
	}
	for _, c := range cases {
		got, err := parseDI(c.addr)
		if err != nil {
			t.Errorf("parseDI(%q) error: %v", c.addr, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseDI(%q) = %x want %x", c.addr, got, c.want)
		}
	}
}

// ========== Unit Convert ==========

func TestUnitConvert(t *testing.T) {
	cases := []struct {
		unit   byte
		value  float64
		expect float64
	}{
		{0x05, 1, 1},    // kWh → ×10^(5-5)=×1
		{0x06, 1, 10},   // kWhx10 → ×10
		{0x07, 1, 100},  // kWhx100
		{0x0B, 1, 1},    // kJ → ×10^(0xB-0xB)=×1
		{0x0E, 1, 1000}, // MJ → ×10^(0xE-0xB)=×1000
		{0x17, 1, 1},    // kW
		{0x2C, 1, 1},    // m3
		{0x35, 1, 1},    // m3/h
		{0x43, 1, 1},    // kJ/h
		{0x01, 1000, 1}, // J → kJ
	}
	for _, c := range cases {
		got := meter.UnitConvert(c.unit, c.value)
		if got != c.expect {
			t.Errorf("unit_convert(0x%02X, %v) = %v want %v", c.unit, c.value, got, c.expect)
		}
	}
}

// ========== Protocol-level: OnData + pending ==========

func TestProtocol_OnData_Match(t *testing.T) {
	c := NewProtocol()

	// 注册物模型 + 挂载
	props := []thing.Property{
		{Key: "volume", Address: "00.01", Format: "XXXXXX.XX", Rate: 1},
	}
	if err := protocol.RegisterModel("model1", &thing.Model{Protocol: "cjt188", Properties: props, Pollers: nil}); err != nil {
		t.Fatal(err)
	}
	if err := c.Mount("dev1", "1", "model1"); err != nil {
		t.Fatal(err)
	}

	// 手动构造一个请求帧，拿到它的 DI/SEQ
	frame := &Frame{
		Type: TypeWater,
		Addr: [7]byte{0, 0, 0, 0, 0, 0, 1},
		Ctrl: CtrlReadData,
		DI:   [2]byte{0x01, 0x00},
		Seq:  0x01,
	}
	reqBytes := frame.Encode()

	// 注册 pending（模拟 Poll 做的事，但不真正 Write）
	c.mu.Lock()
	c.nextSeq() // 推进 seq
	c.mu.Unlock()

	pkey := pendingKey(frame.Addr, frame.Ctrl, frame.DI)
	ch := make(chan *Frame, 1)
	c.pendingMu.Lock()
	c.pending[pkey] = &pendingRequest{key: pkey, ch: ch}
	c.pendingMu.Unlock()

	_ = reqBytes

	// 构造从站响应帧（Ctrl 应答 + DI + payload）
	respFrame := &Frame{
		Type: TypeWater,
		Addr: frame.Addr,
		Ctrl: CtrlReadData | CtrlReplyMask, // 0x81
		DI:   frame.DI,
		Seq:  frame.Seq,
		Data: []byte{0x00, 0x12, 0x34, 0x56},
	}
	respBytes := respFrame.Encode()

	// OnData 喂入
	if _, err := c.OnData(respBytes); err != nil {
		t.Fatalf("OnData 错误: %v", err)
	}

	// 从 channel 取
	select {
	case got := <-ch:
		if got == nil {
			t.Fatal("got nil frame")
		}
		if got.Ctrl != CtrlReadData|CtrlReplyMask {
			t.Errorf("应答 Ctrl 错误: got 0x%02X", got.Ctrl)
		}
		if !bytes.Equal(got.Data, []byte{0x00, 0x12, 0x34, 0x56}) {
			t.Errorf("应答 Data 错误: got %x", got.Data)
		}
	default:
		t.Fatal("OnData 没唤醒 pending channel")
	}
}

func TestProtocol_Close(t *testing.T) {
	c := NewProtocol()
	c.Close() // 不应该 panic
}
