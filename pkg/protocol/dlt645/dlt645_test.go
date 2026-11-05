package dlt645

import (
	"bytes"
	"testing"

	"github.com/god-jason/iot-master/pkg/protocol/meter"
	"github.com/god-jason/iot-master/pkg/thing"
)

// ========== Codec round-trip ==========

func TestCodecRoundTrip_Basic(t *testing.T) {
	// 构造一帧请求：地址 0x010203040506（BCD，低位在前存储：[0x06,0x05,0x04,0x03,0x02,0x01]）
	// CtrlReadData 0x11
	// Data = DI 4 字节 + 可能的 payload，这里只有 DI
	frame := &Frame{
		Addr: [6]byte{0x06, 0x05, 0x04, 0x03, 0x02, 0x01},
		Ctrl: CtrlReadData,
		Data: []byte{0x10, 0x90, 0x00, 0x00}, // DI 0x00009010 (低位在前: 0x10,0x90,0x00,0x00)
	}

	encoded := frame.Encode() // 请求 → +0x33

	// 应该有前导 FEFEFEFE
	if len(encoded) < 4 || encoded[0] != FE || encoded[1] != FE {
		t.Fatalf("前导码错误: %x", encoded[:6])
	}
	// SOH
	if encoded[4] != SOH {
		t.Fatalf("SOH 错误 0x%02X", encoded[4])
	}

	// Decode（模拟从站响应：Data 域已经 sub33）
	// 先做一个从站响应帧
	respFrame := &Frame{
		Addr: frame.Addr,
		Ctrl: CtrlSlaveNormal, // 应答 Ctrl
		Data: []byte{0x10, 0x90, 0x00, 0x00, 0x00, 0x12, 0x34}, // DI + 4字节 BCD 值
	}
	respEncoded := respFrame.Encode() // 响应 → 不 +0x33（构造时已经是 sub33 状态）

	// Decode 应该自动 sub33
	decoded, remaining, err := Decode(respEncoded)
	if err != nil {
		t.Fatalf("Decode 失败: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("剩余数据非空: %x", remaining)
	}

	if decoded.Ctrl != CtrlSlaveNormal {
		t.Errorf("Ctrl 错误: want 0x%02X got 0x%02X", CtrlSlaveNormal, decoded.Ctrl)
	}
	// Data 域应该被 sub33
	for i, v := range decoded.Data {
		expected := respFrame.Data[i]
		if v != expected {
			t.Errorf("Data[%d] 错误: want 0x%02X got 0x%02X (sub33 后)", i, expected, v)
		}
	}
}

func TestCodecRoundTrip_NoPreamble(t *testing.T) {
	// 从站响应经常不带 FEFEFEFE 前导
	frame := &Frame{
		Addr: [6]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		Ctrl: CtrlSlaveNormal,
		Data: []byte{0xAA, 0xBB},
	}
	// Encode 会带前导，手动截取
	full := frame.Encode()
	noFE := full[FECnt:] // 去掉前导

	decoded, _, err := Decode(noFE)
	if err != nil {
		t.Fatalf("无前导码 Decode 失败: %v", err)
	}
	if !bytes.Equal(decoded.Data, []byte{0xAA, 0xBB}) {
		t.Errorf("无前导码 Data 错误: got %x", decoded.Data)
	}
}

func TestCodecRoundTrip_Add33(t *testing.T) {
	// 验证请求 +0x33 和响应 sub33 互逆
	original := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0xAA, 0xFF}
	frame := &Frame{
		Addr: [6]byte{0, 0, 0, 0, 0, 0},
		Ctrl: CtrlReadData,
		Data: original,
	}

	reqEncoded := frame.Encode() // 自动 +0x33
	// 中间 DATA 段位置：FECnt + SOH(1) + ADDR(6) + SOH(1) + CTRL(1) + LEN(1) = FECnt + 10
	dataStart := FECnt + 10
	dataEnd := dataStart + len(original)
	reqDataField := reqEncoded[dataStart:dataEnd]

	for i, v := range reqDataField {
		expected := (original[i] + 0x33) & 0xFF
		if v != byte(expected) {
			t.Errorf("add33[%d] 错误: want 0x%02X got 0x%02X", i, expected, v)
		}
	}

	// Decode 应该自动 sub33 回来
	decoded, _, err := Decode(reqEncoded)
	if err != nil {
		t.Fatalf("Decode 失败: %v", err)
	}
	for i, v := range decoded.Data {
		if v != original[i] {
			t.Errorf("sub33[%d] 错误: want 0x%02X got 0x%02X", i, original[i], v)
		}
	}
}

func TestCodecDecode_Incomplete(t *testing.T) {
	// 截断帧
	frame := &Frame{
		Addr: [6]byte{1, 2, 3, 4, 5, 6},
		Ctrl: CtrlSlaveNormal,
		Data: []byte{0xAA},
	}
	encoded := frame.Encode()
	truncated := encoded[:len(encoded)-1] // 少最后 1 字节

	_, _, err := Decode(truncated)
	if err == nil {
		t.Error("截断帧应该返回错误")
	}
}

func TestCodecDecode_BadChecksum(t *testing.T) {
	frame := &Frame{
		Addr: [6]byte{1, 2, 3, 4, 5, 6},
		Ctrl: CtrlSlaveNormal,
		Data: []byte{0xAA},
	}
	encoded := frame.Encode()
	// 破坏校验
	encoded[len(encoded)-2] ^= 0xFF

	_, _, err := Decode(encoded)
	if err == nil {
		t.Error("校验错误帧应该返回错误")
	}
}

// ========== Mapper / meter.Decode ==========

func TestMeterDecode_BCD(t *testing.T) {
	// meter.lua 里 XXXXXX.XX 格式 → 4 BCD 字节 → rate 0.01
	// BCD 值 0x00 0x12 0x34 0x56 → 00123456 → ×0.01 = 1234.56
	bcd := []byte{0x00, 0x12, 0x34, 0x56}
	v, err := meter.Decode(bcd, "XXXXXX.XX", false, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	fv, ok := v.(float64)
	if !ok {
		t.Fatalf("期望 float64, got %T", v)
	}
	if fv != 1234.56 {
		t.Errorf("BCD XXXXXX.XX 解码错误: want 1234.56 got %v", fv)
	}
}

func TestMeterDecode_BCD_Reverse(t *testing.T) {
	// 逆序场景
	bcd := []byte{0x56, 0x34, 0x12, 0x00}
	v, err := meter.Decode(bcd, "XXXXXX.XX", true, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	fv := v.(float64)
	if fv != 1234.56 {
		t.Errorf("BCD 逆序错误: got %v", fv)
	}
}

func TestMeterDecode_XX(t *testing.T) {
	bcd := []byte{0x25}
	v, err := meter.Decode(bcd, "XX", false, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(float64) != 25 {
		t.Errorf("XX 错误: got %v", v)
	}
}

func TestMeterDecode_XXX_X(t *testing.T) {
	// XXX.X → 2 BCD 字节 → rate 0.1
	bcd := []byte{0x12, 0x34}
	v, err := meter.Decode(bcd, "XXX.X", false, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	fv := v.(float64)
	if fv != 123.4 {
		t.Errorf("XXX.X 错误: want 123.4 got %v", fv)
	}
}

func TestMeterDecode_Datetime6(t *testing.T) {
	// YYMMDDhhmmss → 6 BCD 字节 → "20" + hex
	bcd := []byte{0x24, 0x01, 0x15, 0x10, 0x30, 0x00}
	v, err := meter.Decode(bcd, "YYMMDDhhmmss", false, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "20240115103000" {
		t.Errorf("YYMMDDhhmmss 错误: got %v", v)
	}
}

func TestMeterDecode_Uint16(t *testing.T) {
	b := []byte{0x00, 0x64}
	v, err := meter.Decode(b, "uint16", false, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(float64) != 100 {
		t.Errorf("uint16 错误: got %v", v)
	}
}

func TestDLTMapper_EncodeDecode(t *testing.T) {
	props := []thing.Property{
		{Key: "active_energy", Address: "90100100", Format: "XXXXXX.XX", Rate: 1},
		{Key: "voltage_a", Address: "20000100", Format: "XXX.X", Rate: 1},
	}

	mapper, err := NewMapper(props)
	if err != nil {
		t.Fatal(err)
	}

	// Encode: 每个点独立请求
	reqs := mapper.EncodeRequest()
	if len(reqs) != 2 {
		t.Fatalf("期望 2 个请求 got %d", len(reqs))
	}

	// 第一个请求: DI 0x90100100 低位在前 → [0x00, 0x01, 0x10, 0x90]
	if !bytes.Equal(reqs[0], []byte{0x00, 0x01, 0x10, 0x90}) {
		t.Errorf("DI encode 错误: got %x want 00011090", reqs[0])
	}

	// Decode 响应（已经 sub33 后的原始值）
	// DI 0x90100100 低位在前: [0x00, 0x01, 0x10, 0x90]
	respData := []byte{
		0x00, 0x01, 0x10, 0x90, // DI
		0x00, 0x12, 0x34, 0x56, // XXXXXX.XX BCD: 1234.56
	}

	vals, err := mapper.DecodeResponse(CtrlSlaveNormal, respData)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := vals["active_energy"]; !ok {
		t.Error("缺 active_energy")
	} else if v.(float64) != 1234.56 {
		t.Errorf("active_energy 解码错误: got %v", v)
	}
}

func TestParseDI(t *testing.T) {
	cases := []struct {
		addr string
		want uint32
	}{
		{"90100100", 0x90100100},
		{"0x90100100", 0x90100100},
		{"90.10.01.00", 0x90100100},
		{"20000100", 0x20000100},
	}
	for _, c := range cases {
		got, err := parseDI(c.addr)
		if err != nil {
			t.Errorf("parseDI(%q) error: %v", c.addr, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseDI(%q) = 0x%08X want 0x%08X", c.addr, got, c.want)
		}
	}
}
