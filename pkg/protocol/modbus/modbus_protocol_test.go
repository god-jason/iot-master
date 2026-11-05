package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

// ============================================================
// 测试工具：net.Pipe + bridge goroutine 模拟真实 Modbus 链路
// ============================================================

// startBridge 启动一个 goroutine 不断从 r 读数据，维护拼包缓冲并循环喂给 onData
func startBridge(t *testing.T, r io.Reader, onData func([]byte) (int, error), stopCh chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 256)
		var pending []byte
		for {
			select {
			case <-stopCh:
				return
			default:
			}
			_ = r.(interface{ SetReadDeadline(time.Time) error }).SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := r.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			pending = append(pending, buf[:n]...)
			for {
				consumed, derr := onData(pending)
				_ = derr
				if consumed <= 0 {
					break
				}
				pending = pending[consumed:]
			}
		}
	}()
}

// startBridgeConn 用 net.Conn 版 bridge（维护拼包缓冲）
func startBridgeConn(t *testing.T, conn net.Conn, onData func([]byte) (int, error), stopCh chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 256)
		var pending []byte
		for {
			select {
			case <-stopCh:
				return
			default:
			}
			_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			pending = append(pending, buf[:n]...)
			for {
				consumed, derr := onData(pending)
				_ = derr
				if consumed <= 0 {
					break
				}
				pending = pending[consumed:]
			}
		}
	}()
}

// fakeTCPSlave 读 pipeB 一端的请求，返回 MBAP 响应
type fakeTCPSlave struct {
	t       *testing.T
	conn    net.Conn
	handler func(txID uint16, unitID, fc byte, pdu []byte) []byte
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func newFakeTCPSlave(t *testing.T, conn net.Conn, handler func(txID uint16, unitID, fc byte, pdu []byte) []byte) *fakeTCPSlave {
	s := &fakeTCPSlave{t: t, conn: conn, handler: handler, stopCh: make(chan struct{})}
	s.wg.Add(1)
	go s.loop()
	return s
}

func (s *fakeTCPSlave) loop() {
	defer s.wg.Done()
	buf := make([]byte, 256)
	var pending []byte
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}
		_ = s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := s.conn.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		pending = append(pending, buf[:n]...)

		// 尝试 decode MBAP
		for {
			if len(pending) < 7 {
				break
			}
			h, fc, pdu, rest, derr := DecodeMBAP(pending)
			if derr != nil {
				break
			}
			frameLen := 7 + 1 + len(pdu)
			if len(pending) < frameLen {
				break
			}
			pending = rest

			respPDU := s.handler(h.TransactionID, h.UnitID, fc, pdu)
			if respPDU != nil {
				resp := EncodeMBAP(h.TransactionID, h.UnitID, fc, respPDU)
				_, _ = s.conn.Write(resp)
			}
		}
	}
}

func (s *fakeTCPSlave) Close() {
	close(s.stopCh)
	_ = s.conn.Close()
	s.wg.Wait()
}

// fakeRTUSlave 读 pipeB 一端的 RTU 请求，返回 RTU 响应
type fakeRTUSlave struct {
	t       *testing.T
	conn    net.Conn
	handler func(addr, fc byte, pdu []byte) []byte
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func newFakeRTUSlave(t *testing.T, conn net.Conn, handler func(addr, fc byte, pdu []byte) []byte) *fakeRTUSlave {
	s := &fakeRTUSlave{t: t, conn: conn, handler: handler, stopCh: make(chan struct{})}
	s.wg.Add(1)
	go s.loop()
	return s
}

func (s *fakeRTUSlave) loop() {
	defer s.wg.Done()
	buf := make([]byte, 256)
	var pending []byte
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}
		_ = s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := s.conn.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		pending = append(pending, buf[:n]...)

		for {
			rtu, remaining, derr := DecodeRTU(pending)
			if derr != nil {
				break
			}
			pending = remaining

			respPDU := s.handler(rtu.Addr, rtu.FC, rtu.Data)
			if respPDU != nil {
				resp := EncodeRTU(rtu.Addr, rtu.FC, respPDU)
				_, _ = s.conn.Write(resp)
			}
		}
	}
}

func (s *fakeRTUSlave) Close() {
	close(s.stopCh)
	_ = s.conn.Close()
	s.wg.Wait()
}

// ============================================================
// TCPProtocol 端到端测试
// ============================================================

func TestTCPProtocol_RegisterMountPoll(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()

	handler := func(txID uint16, unitID, fc byte, pdu []byte) []byte {
		startAddr := binary.BigEndian.Uint16(pdu[0:2])
		qty := binary.BigEndian.Uint16(pdu[2:4])
		byteCount := byte(qty * 2)
		resp := make([]byte, 1+byteCount)
		resp[0] = byteCount
		for i := uint16(0); i < qty; i++ {
			binary.BigEndian.PutUint16(resp[1+i*2:3+i*2], startAddr+i+1000)
		}
		return resp
	}
	slave := newFakeTCPSlave(t, pipeB, handler)
	defer slave.Close()

	prop := func(key, addr, dtype string, scale float64) thing.Property {
		return thing.Property{
			Key:       key,
			Address:   addr,
			Format:    dtype,
			Rate:      scale,
			Precision: 1,
		}
	}

	proto := NewTCPProtocol()
	if err := proto.Attach(pipeA, nil); err != nil {
		t.Fatalf("Attach: %v", err)
	}

	// bridge: 读 pipeA → OnData
	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)

	if err := protocol.RegisterModel("m1", &thing.Model{Protocol: "tcp", Properties: []thing.Property{
		prop("voltage", "40001", "uint16", 0.1),
		prop("current", "40002", "int16", 0.01),
	}, Pollers: nil}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := proto.Mount("dev1", "1", "m1"); err != nil {
		t.Fatalf("Mount: %v", err)
	}

	values, err := proto.Poll("dev1", 0)
	close(bridgeStop)
	bridgeWg.Wait()
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}

	// voltage: startAddr=0, 1000*0.1=100.0; current: startAddr=1, 1001*0.01=10.01
	if v, ok := values["voltage"]; !ok {
		t.Error("缺 voltage")
	} else if v.(float64) != 100.0 {
		t.Errorf("voltage: got=%v want=100.0", v)
	}
	if v, ok := values["current"]; !ok {
		t.Error("缺 current")
	} else if v.(float64) != 10.0 { // precision=1 把 10.01 舍入成 10.0
		t.Errorf("current: got=%v want=10.0", v)
	}

	proto.UnMount("dev1")
	protocol.UnregisterModel("model1")
	_ = proto.Close()
}

func TestTCPProtocol_Read(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()

	handler := func(txID uint16, unitID, fc byte, pdu []byte) []byte {
		startAddr := binary.BigEndian.Uint16(pdu[0:2])
		qty := binary.BigEndian.Uint16(pdu[2:4])
		byteCount := byte(qty * 2)
		resp := make([]byte, 1+byteCount)
		resp[0] = byteCount
		for i := uint16(0); i < qty; i++ {
			binary.BigEndian.PutUint16(resp[1+i*2:3+i*2], startAddr+i+50)
		}
		return resp
	}
	slave := newFakeTCPSlave(t, pipeB, handler)
	defer slave.Close()

	proto := NewTCPProtocol()
	_ = proto.Attach(pipeA, nil)

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)

	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "tcp", Properties: []thing.Property{
		{Key: "temp", Address: "40001", Format: "uint16", Rate: 1, Precision: 0},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "1", "m1")

	v, err := proto.Read("dev1", "temp", 0)
	close(bridgeStop)
	bridgeWg.Wait()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if v.(float64) != 50 {
		t.Errorf("Read temp: got=%v want=50", v)
	}
}

func TestTCPProtocol_Write(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()

	// FC06 写单寄存器 → 原样回送 pdu (addr + val = 4 bytes)
	handler := func(txID uint16, unitID, fc byte, pdu []byte) []byte {
		return append([]byte{}, pdu...)
	}
	slave := newFakeTCPSlave(t, pipeB, handler)
	defer slave.Close()

	proto := NewTCPProtocol()
	_ = proto.Attach(pipeA, nil)

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)

	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "tcp", Properties: []thing.Property{
		{Key: "setpoint", Address: "40010", Format: "uint16", Rate: 1},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "1", "m1")

	err := proto.Write("dev1", "setpoint", 200, 0)
	close(bridgeStop)
	bridgeWg.Wait()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestTCPProtocol_Write_NotMounted(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()
	slave := newFakeTCPSlave(t, pipeB, func(txID uint16, unitID, fc byte, pdu []byte) []byte { return nil })
	defer slave.Close()

	proto := NewTCPProtocol()
	_ = proto.Attach(pipeA, nil)
	err := proto.Write("nope", "any", 1, 0)
	if err != protocol.ErrNotMounted {
		t.Errorf("预期 ErrNotMounted, got %v", err)
	}
}

func TestTCPProtocol_Poll_Timeout(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()
	slave := newFakeTCPSlave(t, pipeB, func(txID uint16, unitID, fc byte, pdu []byte) []byte {
		time.Sleep(5 * time.Second)
		return nil
	})
	defer slave.Close()

	proto := NewTCPProtocol()
	proto.timeout = 500 * time.Millisecond
	_ = proto.Attach(pipeA, nil)

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)
	defer func() { close(bridgeStop); bridgeWg.Wait() }()

	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "tcp", Properties: []thing.Property{
		{Key: "v", Address: "40001", Format: "uint16"},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "1", "m1")

	_, err := proto.Poll("dev1", 0)
	if err != protocol.ErrTimeout {
		t.Errorf("预期 ErrTimeout, got %v", err)
	}
}

func TestTCPProtocol_Name(t *testing.T) {
	proto := NewTCPProtocol()
	if proto.Name() != protocol.TCP {
		t.Errorf("Name: got=%s want=%s", proto.Name(), protocol.TCP)
	}
}

func TestTCPProtocol_Close(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()
	slave := newFakeTCPSlave(t, pipeB, func(txID uint16, unitID, fc byte, pdu []byte) []byte { return nil })
	defer slave.Close()

	proto := NewTCPProtocol()
	_ = proto.Attach(pipeA, nil)
	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "tcp", Properties: []thing.Property{
		{Key: "v", Address: "40001", Format: "uint16"},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "1", "m1")
	_ = proto.Close()
}

// ============================================================
// RTUProtocol 端到端测试
// ============================================================

func TestRTUProtocol_RegisterMountPoll(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()

	handler := func(addr, fc byte, pdu []byte) []byte {
		// pdu 格式: [addrHi, addrLo, qtyHi, qtyLo] — 这是请求的 addr+qty
		startAddr := binary.BigEndian.Uint16(pdu[0:2])
		qty := binary.BigEndian.Uint16(pdu[2:4])
		byteCount := byte(qty * 2)
		resp := make([]byte, 1+byteCount)
		resp[0] = byteCount
		for i := uint16(0); i < qty; i++ {
			binary.BigEndian.PutUint16(resp[1+i*2:3+i*2], startAddr+i+2000)
		}
		return resp
	}
	slave := newFakeRTUSlave(t, pipeB, handler)
	defer slave.Close()

	proto := NewRTUProtocol()
	if err := proto.Attach(pipeA, nil); err != nil {
		t.Fatalf("Attach: %v", err)
	}

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)

	if err := protocol.RegisterModel("m1", &thing.Model{Protocol: "rtu", Properties: []thing.Property{
		{Key: "power", Address: "40001", Format: "uint16", Rate: 0.01, Precision: 2},
	}, Pollers: nil}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := proto.Mount("dev1", "2", "m1"); err != nil {
		t.Fatalf("Mount: %v", err)
	}

	values, err := proto.Poll("dev1", 0)
	close(bridgeStop)
	bridgeWg.Wait()
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}

	// power: startAddr=0, 值=2000, *0.01=20.00
	v, ok := values["power"]
	if !ok {
		t.Fatal("缺 power")
	}
	fv := v.(float64)
	if fv < 19.999 || fv > 20.001 {
		t.Errorf("power: got=%v want=20.00", fv)
	}
}

func TestRTUProtocol_Write(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()

	handler := func(addr, fc byte, pdu []byte) []byte {
		// FC06 回显 pdu = [addrHi, addrLo, valHi, valLo]
		return append([]byte{}, pdu...)
	}
	slave := newFakeRTUSlave(t, pipeB, handler)
	defer slave.Close()

	proto := NewRTUProtocol()
	_ = proto.Attach(pipeA, nil)

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)

	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "rtu", Properties: []thing.Property{
		{Key: "valve", Address: "40005", Format: "uint16"},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "3", "m1")

	err := proto.Write("dev1", "valve", 1, 0)
	close(bridgeStop)
	bridgeWg.Wait()
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestRTUProtocol_Poll_Timeout(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()
	slave := newFakeRTUSlave(t, pipeB, func(addr, fc byte, pdu []byte) []byte {
		time.Sleep(5 * time.Second)
		return nil
	})
	defer slave.Close()

	proto := NewRTUProtocol()
	proto.timeout = 500 * time.Millisecond
	_ = proto.Attach(pipeA, nil)

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)
	defer func() { close(bridgeStop); bridgeWg.Wait() }()

	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "rtu", Properties: []thing.Property{
		{Key: "v", Address: "40001", Format: "uint16"},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "1", "m1")

	_, err := proto.Poll("dev1", 0)
	if err != protocol.ErrTimeout {
		t.Errorf("预期 ErrTimeout, got %v", err)
	}
}

func TestRTUProtocol_Name(t *testing.T) {
	proto := NewRTUProtocol()
	if proto.Name() != protocol.RTU {
		t.Errorf("Name: got=%s", proto.Name())
	}
}

// ============================================================
// 工厂注册 / NewProtocol 测试
// ============================================================

func TestFactory_NewTCP(t *testing.T) {
	p, ok := protocol.NewProtocol(protocol.TCP)
	if !ok {
		t.Fatal("TCP factory 未注册")
	}
	if p.Name() != protocol.TCP {
		t.Errorf("Name: got=%s", p.Name())
	}
}

func TestFactory_NewRTU(t *testing.T) {
	p, ok := protocol.NewProtocol(protocol.RTU)
	if !ok {
		t.Fatal("RTU factory 未注册")
	}
	if p.Name() != protocol.RTU {
		t.Errorf("Name: got=%s", p.Name())
	}
}

func TestFactory_Unknown(t *testing.T) {
	_, ok := protocol.NewProtocol("nope")
	if ok {
		t.Error("未知协议不应注册")
	}
}

// ============================================================
// OnData 独立测试：手动喂字节
// ============================================================

func TestTCPProtocol_OnData_Manual(t *testing.T) {
	proto := NewTCPProtocol()
	txID, ch, cancel := proto.registerPending()
	defer cancel()

	// FC3 响应: pdu = [fc, byteCount, data...]
	pdu := []byte{FCReadHoldingRegisters, 4, 0x03, 0xE8, 0x00, 0x64}
	resp := EncodeMBAP(txID, 1, FCReadHoldingRegisters, pdu[1:])

	_, err := proto.OnData(resp)
	if err != nil {
		t.Fatalf("OnData: %v", err)
	}

	select {
	case frame := <-ch:
		if frame == nil {
			t.Fatal("pending channel 返回 nil")
		}
		if frame.IsError {
			t.Fatalf("frame 是 error: %s", frame.ErrorMsg)
		}
		vals := frame.Uint16s()
		if len(vals) != 2 {
			t.Fatalf("Values 数量: got=%d want=2", len(vals))
		}
		if vals[0] != 1000 || vals[1] != 100 {
			t.Errorf("Values: got=%v", vals)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("pending channel 超时未收到")
	}
}

func TestRTUProtocol_OnData_Manual(t *testing.T) {
	proto := NewRTUProtocol()
	ch, cancel := proto.registerPendingRTU(5, FCReadInputRegisters)
	defer cancel()

	// FC4 响应: byteCount=2, data=[0x00, 0x0A] (= 10)
	rtuPDU := []byte{2, 0x00, 0x0A} // byteCount + data (没有 FC！FC 由 RTU 帧头携带)
	resp := EncodeRTU(5, FCReadInputRegisters, rtuPDU)

	_, err := proto.OnData(resp)
	if err != nil {
		t.Fatalf("OnData: %v", err)
	}

	select {
	case frame := <-ch:
		if frame == nil {
			t.Fatal("pending channel 返回 nil")
		}
		if frame.IsError {
			t.Fatalf("frame 是 error: %s", frame.ErrorMsg)
		}
		if vals := frame.Uint16s(); len(vals) != 1 || vals[0] != 10 {
			t.Errorf("Values: got=%v", vals)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("pending channel 超时未收到")
	}
}

// ============================================================
// MBAP / RTU 底层编解码测试
// ============================================================

func TestMBAP_EncodeDecode(t *testing.T) {
	pdu := []byte{FCReadHoldingRegisters, 0x00, 0x00, 0x00, 0x01}
	encoded := EncodeMBAP(42, 1, FCReadHoldingRegisters, pdu[1:])

	h, fc, data, remaining, err := DecodeMBAP(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if h.TransactionID != 42 || h.UnitID != 1 || fc != FCReadHoldingRegisters {
		t.Errorf("header 错误: %+v", h)
	}
	if len(remaining) != 0 {
		t.Errorf("remaining 应为空")
	}
	// data 是 FC 后面的 addr+qty: [00, 00, 00, 01]
	if len(data) != 4 {
		t.Errorf("data len: got=%d want=4", len(data))
	}
}

func TestRTU_EncodeDecode_Request(t *testing.T) {
	// FC3 请求：addr=1, pdu=[addrHi, addrLo, qtyHi, qtyLo] = 4 bytes
	pdu := []byte{0x00, 0x00, 0x00, 0x01}
	encoded := EncodeRTU(1, FCReadHoldingRegisters, pdu)

	rtu, remaining, err := DecodeRTU(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if rtu.Addr != 1 || rtu.FC != FCReadHoldingRegisters {
		t.Errorf("RTU header: addr=%d fc=%02X", rtu.Addr, rtu.FC)
	}
	if len(remaining) != 0 {
		t.Errorf("remaining 应为空")
	}
	if len(rtu.Data) != 4 {
		t.Errorf("Data len: got=%d want=4", len(rtu.Data))
	}
}

func TestRTU_EncodeDecode_ReadResponse(t *testing.T) {
	// FC3 读响应: byteCount=4, data=[0x00, 0x64, 0x00, 0xC8] = [100, 200]
	pdu := []byte{4, 0x00, 0x64, 0x00, 0xC8}
	encoded := EncodeRTU(1, FCReadHoldingRegisters, pdu)

	rtu, remaining, err := DecodeRTU(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(rtu.Data) != 5 {
		t.Errorf("Data len: got=%d want=5", len(rtu.Data))
	}
	if rtu.Data[0] != 4 {
		t.Errorf("byteCount: got=%d want=4", rtu.Data[0])
	}
	if len(remaining) != 0 {
		t.Errorf("remaining 应为空")
	}
}

func TestRTU_CRCError(t *testing.T) {
	pdu := []byte{0x00, 0x00, 0x00, 0x01}
	encoded := EncodeRTU(1, FCReadHoldingRegisters, pdu)
	encoded[0] = 0x99 // 改 addr 让 CRC 错

	_, _, err := DecodeRTU(encoded)
	if err != ErrCRC {
		t.Errorf("预期 ErrCRC, got %v", err)
	}
}

func TestRTU_ExceptionResponse(t *testing.T) {
	// FC3 exception: fc|0x80, excCode=0x02
	pdu := []byte{0x02}
	encoded := EncodeRTU(1, FCReadHoldingRegisters|0x80, pdu)

	rtu, remaining, err := DecodeRTU(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if rtu.FC != FCReadHoldingRegisters|0x80 {
		t.Errorf("FC should be exception")
	}
	if len(remaining) != 0 {
		t.Errorf("remaining 应为空")
	}
}

func TestBuildReadRequest(t *testing.T) {
	req := BuildReadRequest(100, 5)
	if len(req) != 4 {
		t.Fatalf("len: got=%d want=4", len(req))
	}
	if binary.BigEndian.Uint16(req[0:2]) != 100 {
		t.Errorf("startAddr: got=%d", binary.BigEndian.Uint16(req[0:2]))
	}
	if binary.BigEndian.Uint16(req[2:4]) != 5 {
		t.Errorf("qty: got=%d", binary.BigEndian.Uint16(req[2:4]))
	}
}

func TestParseReadResponse(t *testing.T) {
	values, err := ParseReadResponse(FCReadHoldingRegisters, []byte{4, 0x00, 0x64, 0x00, 0xC8})
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0] != 100 || values[1] != 200 {
		t.Errorf("values: %v", values)
	}

	_, err = ParseReadResponse(FCReadHoldingRegisters|0x80, []byte{0x02})
	if err == nil {
		t.Error("预期 exception 错误")
	}
}

func TestParseAddr(t *testing.T) {
	spec, err := ParseAddr("40001", TypeUint16, 0)
	if err != nil {
		t.Fatal(err)
	}
	if spec.FC != FCReadHoldingRegisters || spec.StartAddr != 0 || spec.Quantity != 1 {
		t.Errorf("40001: fc=%02X addr=%d qty=%d", spec.FC, spec.StartAddr, spec.Quantity)
	}

	spec, err = ParseAddr("FC3:100:10", TypeUint16, 0)
	if err != nil {
		t.Fatal(err)
	}
	if spec.FC != FCReadHoldingRegisters || spec.StartAddr != 100 || spec.Quantity != 10 {
		t.Errorf("FC3:100:10: fc=%02X addr=%d qty=%d", spec.FC, spec.StartAddr, spec.Quantity)
	}

	// 30005 数字前缀推断 FC4 input
	spec, err = ParseAddr("30005", TypeUint16, 0)
	if err != nil {
		t.Fatal(err)
	}
	if spec.FC != FCReadInputRegisters || spec.StartAddr != 4 {
		t.Errorf("30005: fc=%02X addr=%d want addr=4", spec.FC, spec.StartAddr)
	}

	// 5 在 1-9999 区间 → FC1 coil
	spec, err = ParseAddr("5", TypeUint16, 0)
	if err != nil {
		t.Fatal(err)
	}
	if spec.FC != FCReadCoils {
		t.Errorf("5 → fc=%02X want FC1", spec.FC)
	}
	if spec.StartAddr != 4 {
		t.Errorf("addr 5 → start=%d want=4", spec.StartAddr)
	}

	// FC2 显式 + 数字地址
	spec, err = ParseAddr("FC2:10", TypeUint16, 0)
	if err != nil {
		t.Fatal(err)
	}
	if spec.FC != FCReadDiscreteInputs || spec.StartAddr != 10 {
		t.Errorf("FC2:10: fc=%02X addr=%d", spec.FC, spec.StartAddr)
	}
}

func TestParseAddrByte(t *testing.T) {
	if v, _ := parseAddrOK("1"); v != 1 {
		t.Error("parseAddrOK 1 fail")
	}
	if v, _ := parseAddrOK("abc"); v != 0 {
		t.Error("parseAddrOK non-numeric should be 0")
	}
}

func TestToUint16(t *testing.T) {
	if v, ok := toUint16(float64(200)); !ok || v != 200 {
		t.Errorf("float64→uint16: got %d ok=%v", v, ok)
	}
	if v, ok := toUint16(int(65535)); !ok || v != 65535 {
		t.Errorf("int→uint16: got %d ok=%v", v, ok)
	}
	if v, ok := toUint16("nope"); ok || v != 0 {
		t.Errorf("unknown type: got %d ok=%v want (0, false)", v, ok)
	}
}

func TestCRC16Modbus_Vector(t *testing.T) {
	// 已知向量: addr=1, fc=3, startAddr=0, qty=1 → crc 低字节在前
	// 手工计算: data = [01, 03, 00, 00, 00, 01]
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x01}
	crc := crc16Modbus(data)

	// EncodeRTU 会算同样的 CRC
	// EncodeRTU(1, 3, [00, 00, 00, 01]) 应该有相同的 CRC
	rtu := EncodeRTU(1, FCReadHoldingRegisters, []byte{0x00, 0x00, 0x00, 0x01})
	// rtu = [addr=01, fc=03, pdu=00 00 00 01, crc_low, crc_high]
	expected := binary.LittleEndian.Uint16(rtu[6:8])
	if crc != expected {
		t.Errorf("CRC mismatch: crc16=%04X encodeRTU=%04X", crc, expected)
	}
}

func TestBuildFrameFromPdu_Error(t *testing.T) {
	f := BuildFrameFromPdu(protocol.TCP, 0, 1, FCReadHoldingRegisters|0x80, []byte{0x02}, nil)
	if !f.IsError {
		t.Error("预期 IsError=true")
	}
}

func TestProtocolInterface(t *testing.T) {
	var _ protocol.Protocol = NewTCPProtocol()
	var _ protocol.Protocol = NewRTUProtocol()
}

func TestFindCachedForDevice(t *testing.T) {
	protocol.RegisterModel("m1", &thing.Model{Protocol: "tcp", Properties: []thing.Property{
		{Key: "v", Address: "40001", Format: "uint16"},
	}, Pollers: nil})

	cachedAny, ok := protocol.GetModel("m1")
	if !ok {
		t.Fatal("GetModel 失败")
	}
	cached, ok := cachedAny.(*ModbusModel)
	if !ok {
		t.Fatal("类型断言失败")
	}
	if len(cached.Items) != 1 {
		t.Errorf("Items: got=%d want=1", len(cached.Items))
	}
}

func TestClose_Double(t *testing.T) {
	proto := NewTCPProtocol()
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()
	slave := newFakeTCPSlave(t, pipeB, func(txID uint16, unitID, fc byte, pdu []byte) []byte { return nil })
	defer slave.Close()
	_ = proto.Attach(pipeA, nil)
	_ = proto.Close()
	_ = proto.Close() // 不 panic
}

// TestRTUPoll_MultipleMappers 验证 RTU 连续读取多 Mapper 并按 FC 合并
func TestRTUPoll_MultipleMappers(t *testing.T) {
	pipeA, pipeB := net.Pipe()
	defer pipeA.Close()

	callCount := 0
	handler := func(addr, fc byte, pdu []byte) []byte {
		callCount++
		startAddr := binary.BigEndian.Uint16(pdu[0:2])
		qty := binary.BigEndian.Uint16(pdu[2:4])
		byteCount := byte(qty * 2)
		resp := make([]byte, 1+byteCount)
		resp[0] = byteCount
		for i := uint16(0); i < qty; i++ {
			binary.BigEndian.PutUint16(resp[1+i*2:3+i*2], startAddr+i+3000)
		}
		return resp
	}
	slave := newFakeRTUSlave(t, pipeB, handler)
	defer slave.Close()

	proto := NewRTUProtocol()
	_ = proto.Attach(pipeA, nil)

	var bridgeWg sync.WaitGroup
	bridgeStop := make(chan struct{})
	startBridgeConn(t, pipeA, proto.OnData, bridgeStop, &bridgeWg)

	_ = protocol.RegisterModel("m1", &thing.Model{Protocol: "rtu", Properties: []thing.Property{
		{Key: "a", Address: "40001", Format: "uint16", Rate: 1},
		{Key: "b", Address: "40002", Format: "uint16", Rate: 1},
		{Key: "c", Address: "30001", Format: "uint16", Rate: 1},
	}, Pollers: nil})
	_ = proto.Mount("dev1", "1", "m1")

	values, err := proto.Poll("dev1", 0)
	close(bridgeStop)
	bridgeWg.Wait()
	if err != nil {
		t.Fatal(err)
	}
	// FC3 合并 1 次（a,b）, FC4 1 次（c） → 共 2 次
	if callCount != 2 {
		t.Errorf("handler 调用次数: got=%d want=2", callCount)
	}
	if len(values) != 3 {
		t.Errorf("values 数量: got=%d want=3", len(values))
	}
	fmt.Printf("values: %+v\n", values)
}
