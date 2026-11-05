// gatewaytest 软网关主动轮询测试程序（无需 DB）。
//
// 本进程同时扮演两个角色：
//
//	DTU      —— 连接软网关 Server，发送文本注册包（SN），然后等待网关下发 Modbus 请求
//	Modbus   —— 在同一 TCP 连接上作为虚拟从站，解析 MBAP 读/写请求并回写响应
//
// 网关侧通过自定义 ProtocolHandler 在链路注册时直接 Mount 一个虚拟设备（不查 DB），
// 物模型由 iot.SetModel 注入，随后由 Poller 按 -interval 主动轮询，
// 最后校验设备属性被轮询结果填充。
//
// 用法：
//
//	go run ./cmd/gatewaytest
//	go run ./cmd/gatewaytest -listen=127.0.0.1:18001 -interval=500ms
package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/god-jason/iot-master/internal/gateway"
	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/protocol/modbus"
	"github.com/god-jason/iot-master/pkg/thing"
)

var (
	listen   = flag.String("listen", "127.0.0.1:18001", "软网关 TCP 监听地址")
	sn       = flag.String("sn", "GW-TEST-0001", "模拟 DTU 注册包 SN")
	modelID  = flag.String("model", "modbus-test-model", "虚拟设备物模型 ID")
	deviceID = flag.String("device", "modbus-test-device", "虚拟设备 ID")
	address  = flag.String("address", "1", "虚拟设备从站地址")
	interval = flag.Duration("interval", time.Second, "网关轮询间隔")
	waitSec  = flag.Int("wait", 10, "等待轮询结果的超时（秒）")
)

// 虚拟从站保持寄存器初值：temperature=25.6（Rate=0.1）、humidity=61.5（Rate=0.1）
const (
	rawTemperature uint16 = 256
	rawHumidity    uint16 = 615
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Printf("\n测试失败: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	logger.Init("info")

	// 1. 注册物模型（iot 缓存 + 协议归一化，均不落库）
	model := &thing.Model{
		Protocol: protocol.TCP,
		Properties: []thing.Property{
			{Key: "temperature", Name: "温度", Type: "float", Mode: "r", Unit: "℃",
				Register: "holding_register", Address: "40001", Format: "uint16", Rate: 0.1, Precision: 1},
			{Key: "humidity", Name: "湿度", Type: "float", Mode: "r", Unit: "%",
				Register: "holding_register", Address: "40002", Format: "uint16", Rate: 0.1, Precision: 1},
		},
	}
	iot.SetModel(*modelID, model)

	// 2. 启动软网关 Server
	handler := &mountHandler{
		device: &types.Device{
			ID:      *deviceID,
			SN:      *deviceID,
			Name:    "虚拟 Modbus 设备",
			ModelID: *modelID,
			Address: *address,
		},
	}
	srv := gateway.NewServer(
		gateway.WithServerID(1),
		gateway.WithDefaultProtocol(protocol.TCP),
		gateway.WithHandler(handler),
		gateway.WithConcurrency(1),
		gateway.WithTickInterval(100*time.Millisecond),
		gateway.WithInterval(*interval),
	)
	handler.srv = srv
	defer srv.Stop()

	go func() {
		if err := srv.Listen("tcp://" + *listen); err != nil {
			fmt.Printf("[gateway] Server 退出: %v\n", err)
		}
	}()
	if err := waitTCP(*listen, 3*time.Second); err != nil {
		return err
	}
	fmt.Printf("[gateway] 已监听 %s，轮询间隔 %s\n", *listen, *interval)

	// 3. 模拟 DTU 连接并注册
	conn, err := net.DialTimeout("tcp", *listen, 3*time.Second)
	if err != nil {
		return fmt.Errorf("连接软网关失败: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(*sn + "\n")); err != nil {
		return fmt.Errorf("发送注册包失败: %w", err)
	}
	fmt.Printf("[dtu] 已连接并发送注册包 %q\n", *sn)

	// 4. 同一连接上作为虚拟 Modbus TCP 从站应答网关请求
	slv := newSlave()
	go func() {
		if err := slv.serve(conn); err != nil {
			fmt.Printf("[slave] 应答结束: %v\n", err)
		}
	}()

	// 5. 等待设备注册到 iot
	dev, err := waitDevice(*deviceID, time.Duration(*waitSec)*time.Second)
	if err != nil {
		return err
	}

	// 6. 等待首次轮询结果
	timeout := time.Duration(*waitSec) * time.Second
	deadline := time.Now().Add(timeout)
	var props map[string]interface{}
	for {
		props = dev.Properties()
		if _, ok := props["temperature"]; ok {
			if _, ok := props["humidity"]; ok {
				break
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("等待轮询结果超时（%s），从站已收到 %d 次读请求", timeout, slv.reads())
		}
		time.Sleep(50 * time.Millisecond)
	}

	temperature, _ := floatProp(props, "temperature")
	humidity, _ := floatProp(props, "humidity")
	fmt.Printf("[gateway] 首次轮询结果: temperature=%v humidity=%v（从站读请求 %d 次）\n",
		temperature, humidity, slv.reads())

	if !near(temperature, float64(rawTemperature)*0.1) {
		return fmt.Errorf("temperature 期望 %v，实际 %v", float64(rawTemperature)*0.1, temperature)
	}
	if !near(humidity, float64(rawHumidity)*0.1) {
		return fmt.Errorf("humidity 期望 %v，实际 %v", float64(rawHumidity)*0.1, humidity)
	}

	// 7. 再等待一轮，确认网关在持续主动轮询
	before := slv.reads()
	time.Sleep(*interval + 500*time.Millisecond)
	after := slv.reads()
	if after <= before {
		return fmt.Errorf("轮询未持续：等待 %s 后读请求数量未增加（%d → %d）", *interval+500*time.Millisecond, before, after)
	}
	fmt.Printf("[gateway] 持续轮询正常：%s 内读请求 %d → %d 次\n", *interval+500*time.Millisecond, before, after)

	fmt.Println("\n软网关主动轮询测试通过")
	return nil
}

// ============ 软网关业务回调 ============

type mountHandler struct {
	srv    *gateway.Server
	device *types.Device
}

// OnRegister 链路注册：直接挂载虚拟设备并注册到 iot（不访问 DB）
func (h *mountHandler) OnRegister(_ context.Context, link *gateway.Link) {
	devs := h.srv.MountDevices(link, []*types.Device{h.device})
	for _, dev := range devs {
		dev.SetOnline()
		iot.Register(dev.ID(), dev)
	}
	fmt.Printf("[gateway] 链路 %s 注册完成，已挂载 %d 个设备\n", link.ID(), len(devs))
}

// OnDisconnect 链路断开：注销 iot 设备
func (h *mountHandler) OnDisconnect(_ context.Context, link *gateway.Link, reason string) {
	iot.Unregister(h.device.ID)
	fmt.Printf("[gateway] 链路 %s 断开: %s\n", link.ID(), reason)
}

// ============ 虚拟 Modbus TCP 从站 ============

type slave struct {
	mu     sync.Mutex
	regs   []uint16
	readN  int
	writeN int
}

func newSlave() *slave {
	return &slave{regs: []uint16{rawTemperature, rawHumidity}}
}

func (s *slave) reads() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readN
}

// serve 循环读取 MBAP 请求并回写响应，直到连接断开
func (s *slave) serve(conn net.Conn) error {
	r := bufio.NewReader(conn)
	for {
		hdr := make([]byte, 7) // txID(2) + protocolID(2) + length(2) + unitID(1)
		if _, err := io.ReadFull(r, hdr); err != nil {
			return err
		}
		txID := binary.BigEndian.Uint16(hdr[0:2])
		length := int(binary.BigEndian.Uint16(hdr[4:6]))
		unitID := hdr[6]
		if length < 2 || length > 260 {
			return fmt.Errorf("MBAP length 非法: %d", length)
		}

		body := make([]byte, length-1) // FC + PDU
		if _, err := io.ReadFull(r, body); err != nil {
			return err
		}
		fc, pdu := body[0], body[1:]

		respFC, respPDU := s.handle(fc, pdu)
		if _, err := conn.Write(modbus.EncodeMBAP(txID, unitID, respFC, respPDU)); err != nil {
			return err
		}
	}
}

// handle 处理单个请求 PDU（含 FC），返回响应 FC 与响应 PDU（不含 FC）
func (s *slave) handle(fc byte, pdu []byte) (byte, []byte) {
	switch fc {
	case modbus.FCReadHoldingRegisters, modbus.FCReadInputRegisters:
		if len(pdu) < 4 {
			return fc | 0x80, []byte{0x03}
		}
		start := binary.BigEndian.Uint16(pdu[0:2])
		qty := binary.BigEndian.Uint16(pdu[2:4])
		vals, ok := s.readRegs(start, qty)
		if !ok {
			return fc | 0x80, []byte{0x02} // 非法数据地址
		}
		out := make([]byte, 1+2*len(vals))
		out[0] = byte(2 * len(vals))
		for i, v := range vals {
			binary.BigEndian.PutUint16(out[1+2*i:], v)
		}
		return fc, out

	case modbus.FCWriteSingleRegister:
		if len(pdu) < 4 {
			return fc | 0x80, []byte{0x03}
		}
		if !s.writeRegs(binary.BigEndian.Uint16(pdu[0:2]), []uint16{binary.BigEndian.Uint16(pdu[2:4])}) {
			return fc | 0x80, []byte{0x02}
		}
		return fc, append([]byte(nil), pdu...)

	case modbus.FCWriteMultipleRegs:
		if len(pdu) < 5 {
			return fc | 0x80, []byte{0x03}
		}
		start := binary.BigEndian.Uint16(pdu[0:2])
		qty := int(binary.BigEndian.Uint16(pdu[2:4]))
		if len(pdu) < 5+2*qty {
			return fc | 0x80, []byte{0x03}
		}
		vals := make([]uint16, qty)
		for i := 0; i < qty; i++ {
			vals[i] = binary.BigEndian.Uint16(pdu[5+2*i:])
		}
		if !s.writeRegs(start, vals) {
			return fc | 0x80, []byte{0x02}
		}
		return fc, append([]byte(nil), pdu[:4]...)

	default:
		return fc | 0x80, []byte{0x01} // 非法功能码
	}
}

func (s *slave) readRegs(start, qty uint16) ([]uint16, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if qty == 0 || int(start)+int(qty) > len(s.regs) {
		return nil, false
	}
	s.readN++
	return append([]uint16(nil), s.regs[start:int(start)+int(qty)]...), true
}

func (s *slave) writeRegs(start uint16, vals []uint16) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(vals) == 0 || int(start)+len(vals) > len(s.regs) {
		return false
	}
	copy(s.regs[start:], vals)
	s.writeN++
	return true
}

// ============ 工具 ============

func waitTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return nil
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("等待 %s 就绪超时: %v", addr, lastErr)
}

func waitDevice(id string, timeout time.Duration) (iot.Device, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if d, ok := iot.GetDevice(id); ok {
			return d, nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil, fmt.Errorf("设备 %s 未在 %s 内注册到 iot", id, timeout)
}

func floatProp(props map[string]interface{}, key string) (float64, bool) {
	v, ok := props[key]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	}
	return 0, false
}

func near(v, want float64) bool {
	d := v - want
	return d > -0.05 && d < 0.05
}
