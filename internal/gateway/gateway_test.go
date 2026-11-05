package gateway

import (
	"io"
	"testing"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/protocol"
	"github.com/god-jason/iot-master/pkg/thing"
)

func TestParseInterval(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"5s", 5 * time.Second},
		{"1m30s", 90 * time.Second},
		{"", 0},
		{"bad", 0},
		{"-1s", 0},
		{"0s", 0},
	}
	for _, c := range cases {
		if got := parseInterval(c.in); got != c.want {
			t.Errorf("parseInterval(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestReadTextFrame(t *testing.T) {
	line, consumed, ok := readTextFrame([]byte("GW001\r\nrest"))
	if !ok || string(line) != "GW001" || consumed != 6 {
		t.Errorf("readTextFrame = (%q, %d, %v)", line, consumed, ok)
	}

	// 二进制首字节 → 视为未完成/非文本
	if _, _, ok := readTextFrame([]byte{0x01, 0x03, 0x00}); ok {
		t.Error("二进制帧不应识别为文本行")
	}

	// 无行尾且未达上限 → 继续等待
	if _, _, ok := readTextFrame([]byte("GW0")); ok {
		t.Error("不足一行应返回 ok=false")
	}
}

func TestReadLineFrame(t *testing.T) {
	line, consumed, ok := readLineFrame([]byte(" ping \nzzz"))
	if !ok || string(line) != "ping" || consumed != 7 {
		t.Errorf("readLineFrame = (%q, %d, %v)", line, consumed, ok)
	}
	if _, _, ok := readLineFrame([]byte("ping")); ok {
		t.Error("无行尾不应识别为行")
	}
}

func TestIsHeartbeat(t *testing.T) {
	for _, s := range []string{"heartbeat", " HB ", "ping", "PONG", "keepalive", "ka"} {
		if !isHeartbeat(s) {
			t.Errorf("isHeartbeat(%q) 应为 true", s)
		}
	}
	for _, s := range []string{"", "hello", "GW001"} {
		if isHeartbeat(s) {
			t.Errorf("isHeartbeat(%q) 应为 false", s)
		}
	}
}

func TestLooksLikeModbusFrame(t *testing.T) {
	if !looksLikeModbusFrame([]byte{0x01, 0x03, 0x00}) {
		t.Error("0x01/0x03 应识别为 Modbus 帧")
	}
	if looksLikeModbusFrame([]byte{0x01}) {
		t.Error("过短不应识别")
	}
	if looksLikeModbusFrame([]byte{0xF8, 0x03}) {
		t.Error("非法从站地址不应识别")
	}
	if looksLikeModbusFrame([]byte{0x01, 0x18}) {
		t.Error("非法功能码不应识别")
	}
}

func TestDecodeJSONPacket(t *testing.T) {
	pkt, consumed, ok, err := decodeJSONPacket([]byte(`{"sn":"GW1","imei":"86"}}extra`))
	if err != nil || !ok {
		t.Fatalf("decodeJSONPacket: ok=%v err=%v", ok, err)
	}
	if pkt.SN != "GW1" || pkt.IMEI != "86" {
		t.Errorf("pkt = %+v", pkt)
	}
	if consumed <= 0 || consumed > len([]byte(`{"sn":"GW1","imei":"86"}}extra`)) {
		t.Errorf("consumed = %d", consumed)
	}

	if _, consumed, ok, err := decodeJSONPacket([]byte(`{"sn":"GW1"`)); ok || err != nil || consumed != 0 {
		t.Errorf("不完整 JSON 应 ok=false, got consumed=%d ok=%v err=%v", consumed, ok, err)
	}

	if _, _, ok, err := decodeJSONPacket([]byte(`{"sn":}`)); ok || err == nil {
		t.Errorf("非法 JSON 应返回错误, got ok=%v err=%v", ok, err)
	}
}

func TestConsumeRegistrationText(t *testing.T) {
	s := NewServer()
	l := &Link{ConnID: 1}
	consumed, registered, err := s.consumeRegistration(l, []byte("GW001\r\n"))
	if err != nil || !registered || consumed != 6 {
		t.Fatalf("consumed=%d registered=%v err=%v", consumed, registered, err)
	}
	if l.ID() != "GW001" || l.SN() != "GW001" || !l.IsRegistered() {
		t.Errorf("链路身份不正确: id=%q sn=%q registered=%v", l.ID(), l.SN(), l.IsRegistered())
	}
}

func TestConsumeRegistrationJSON(t *testing.T) {
	s := NewServer()
	l := &Link{ConnID: 2}
	consumed, registered, err := s.consumeRegistration(l, []byte(`{"sn":"GW002","imei":"8600001","name":"网关2"}`))
	if err != nil || !registered || consumed != len(`{"sn":"GW002","imei":"8600001","name":"网关2"}`) {
		t.Fatalf("consumed=%d registered=%v err=%v", consumed, registered, err)
	}
	if l.ID() != "GW002" || l.IMEI() != "8600001" || l.Name() != "网关2" {
		t.Errorf("链路身份不正确: %+v", l)
	}
}

func TestConsumeRegistrationHeartbeatIgnored(t *testing.T) {
	s := NewServer()
	l := &Link{ConnID: 3}
	consumed, registered, err := s.consumeRegistration(l, []byte("ping\n"))
	if err != nil || registered || consumed == 0 {
		t.Errorf("心跳包应消费但不算注册: consumed=%d registered=%v err=%v", consumed, registered, err)
	}
	if l.IsRegistered() {
		t.Error("心跳包不应完成注册")
	}
}

func TestConsumeRegistrationBinaryRejected(t *testing.T) {
	s := NewServer()
	l := &Link{ConnID: 4}
	if _, _, err := s.consumeRegistration(l, []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x01}); err != errUnregisteredFrame {
		t.Errorf("未注册连接发协议帧应返回 errUnregisteredFrame, got %v", err)
	}
}

func TestLinkRegisterPriority(t *testing.T) {
	cases := []struct {
		pkt  RegisterPacket
		want string
	}{
		{RegisterPacket{SN: "SN1", IMEI: "IMEI1", Device: "D1"}, "SN1"},
		{RegisterPacket{IMEI: "IMEI1", Device: "D1"}, "IMEI1"},
		{RegisterPacket{Device: "D1"}, "D1"},
		{RegisterPacket{}, ""},
	}
	for _, c := range cases {
		l := &Link{}
		l.register(c.pkt)
		if l.ID() != c.want {
			t.Errorf("register(%+v).ID() = %q, want %q", c.pkt, l.ID(), c.want)
		}
	}
}

func TestLinkDevicesSnapshot(t *testing.T) {
	d1 := &Device{}
	l := &Link{devices: []*Device{d1}}

	got := l.Devices()
	if len(got) != 1 || got[0] != d1 {
		t.Fatalf("Devices() = %v", got)
	}
	got[0] = &Device{}
	if l.Devices()[0] != d1 {
		t.Error("Devices() 应返回副本，修改返回值不影响内部列表")
	}

	if copyDevices(nil) != nil {
		t.Error("copyDevices(nil) 应返回 nil")
	}
}

func TestTouchAndOffline(t *testing.T) {
	link := &Link{}
	dev := newDevice(link, &types.Device{ID: "d1", SN: "SN1", ModelID: "m1"})

	if !dev.OnlineStatus() {
		t.Fatal("newDevice 应默认在线")
	}

	link.replaceDevices([]*Device{dev})
	now := time.Now()
	link.touch(now)
	if !dev.OnlineStatus() || !dev.LastSeenAt().Equal(now) {
		t.Errorf("touch 后设备应在线且 lastSeen 更新")
	}

	link.setOffline()
	if dev.OnlineStatus() {
		t.Error("setOffline 后设备应离线")
	}
}

func TestMountDevices(t *testing.T) {
	const modelID = "mb-mount-model"
	if err := protocol.RegisterModel(modelID, &thing.Model{
		Protocol: protocol.TCP,
		Properties: []thing.Property{
			{Key: "temp", Register: "holding_register", Address: "40001", Format: "uint16", Rate: 0.01},
		},
	}); err != nil {
		t.Fatalf("RegisterModel: %v", err)
	}

	proto, ok := protocol.NewProtocol(protocol.TCP)
	if !ok {
		t.Fatal("Modbus TCP 协议未注册")
	}
	link := &Link{ServerID: 1, ConnID: 1, Proto: proto, Protocol: protocol.TCP}
	s := NewServer()

	rows := []*types.Device{
		{ID: "d1", SN: "SN1", ModelID: modelID, Address: "1"},
		{ID: "d2", ModelID: ""},                              // 无模型 → 跳过
		{ID: "d3", ModelID: modelID, Disabled: true},         // 禁用 → 跳过
		{ID: "d4", ModelID: modelID, Address: "bad-address"}, // 非法地址 → Mount 失败
	}
	devs := s.MountDevices(link, rows)
	if len(devs) != 1 {
		t.Fatalf("MountDevices 挂载数量 = %d, want 1", len(devs))
	}
	if devs[0].ID() != "d1" || !devs[0].Mounted() {
		t.Errorf("设备未正确挂载: %+v", devs[0])
	}
	if got := link.Devices(); len(got) != 1 || got[0] != devs[0] {
		t.Errorf("link.Devices 与挂载结果不一致")
	}

	// 再次挂载应替换旧设备，旧设备 UnMount
	devs2 := s.MountDevices(link, rows)
	if len(devs2) != 1 {
		t.Fatalf("二次挂载数量 = %d", len(devs2))
	}
	if devs2[0] == devs[0] {
		t.Error("二次挂载应创建新的运行时设备")
	}
}

func TestServerAccessors(t *testing.T) {
	s := NewServer(WithServerID(9), WithInterval(time.Second))
	l := &Link{ConnID: 21, ServerID: 9}
	s.links[21] = l

	if _, ok := s.GetLink(21); !ok {
		t.Fatal("GetLink 应找到链路")
	}
	if len(s.Links()) != 1 {
		t.Errorf("Links() = %d", len(s.Links()))
	}
	if s.SchedulerEntries() != 1 {
		t.Errorf("SchedulerEntries() = %d", s.SchedulerEntries())
	}

	dev := newDevice(l, &types.Device{ID: "d1", ModelID: "m1"})
	l.replaceDevices([]*Device{dev})
	if len(s.Devices()) != 1 {
		t.Errorf("Devices() = %d", len(s.Devices()))
	}

	s.SetPoll(21, &PollConfig{Interval: 3 * time.Second})
	if pc := l.PollConfig(); pc == nil || pc.Interval != 3*time.Second {
		t.Errorf("SetPoll 未生效: %+v", pc)
	}

	if _, ok := s.GetLink(999); ok {
		t.Error("不存在的 connID 不应命中")
	}
}

// ===== 轮询派发 =====

type fakeProtocol struct {
	polled []string
}

func (f *fakeProtocol) Name() string { return "fake" }
func (f *fakeProtocol) Attach(io.ReadWriteCloser, protocol.OnValues) error {
	return nil
}
func (f *fakeProtocol) Close() error                       { return nil }
func (f *fakeProtocol) Mount(string, string, string) error { return nil }
func (f *fakeProtocol) UnMount(string)                     {}
func (f *fakeProtocol) Poll(deviceID string, _ time.Duration) (map[string]any, error) {
	f.polled = append(f.polled, deviceID)
	return map[string]any{"v": 1}, nil
}
func (f *fakeProtocol) Read(string, string, time.Duration) (any, error) { return nil, nil }
func (f *fakeProtocol) Write(string, string, any, time.Duration) error  { return nil }
func (f *fakeProtocol) OnData([]byte) (int, error)                      { return 0, nil }

func TestPollerDispatchAndPollOnce(t *testing.T) {
	// 预置物模型缓存，避免 pollOnce 走 DB 加载
	iot.SetModel("poll-model", &thing.Model{})
	defer iot.DeleteModel("poll-model")

	proto := &fakeProtocol{}
	s := NewServer(WithInterval(time.Millisecond))
	l := &Link{
		ConnID:     1,
		Proto:      proto,
		Protocol:   "fake",
		Registered: true,
		lastPoll:   time.Now().Add(-time.Hour),
	}
	dev := newDevice(l, &types.Device{ID: "d1", ModelID: "poll-model"})
	dev.setMounted(true)
	l.devices = []*Device{dev}
	s.links[1] = l

	p := NewPoller(s, 1, time.Millisecond, time.Millisecond)
	p.tickRound()

	select {
	case job := <-p.jobCh:
		if len(job.targets) != 1 || job.targets[0] != dev {
			t.Fatalf("派发目标不正确: %+v", job.targets)
		}
		p.pollOnce(job)
	default:
		t.Fatal("到期链路未派发轮询任务")
	}

	if len(proto.polled) != 1 || proto.polled[0] != "d1" {
		t.Errorf("协议 Poll 调用记录 = %v", proto.polled)
	}
	if v, ok := dev.Property("v"); !ok || v != 1 {
		t.Errorf("轮询结果未写入设备属性: %v (ok=%v)", v, ok)
	}
	if l.busy {
		t.Error("pollOnce 应结束 busy 状态")
	}
}

func TestPollerSkipsUnregisteredLink(t *testing.T) {
	proto := &fakeProtocol{}
	s := NewServer(WithInterval(time.Millisecond))
	l := &Link{ConnID: 2, Proto: proto, lastPoll: time.Now().Add(-time.Hour)}
	dev := newDevice(l, &types.Device{ID: "d1"})
	dev.setMounted(true)
	l.devices = []*Device{dev}
	s.links[2] = l

	p := NewPoller(s, 1, time.Millisecond, time.Millisecond)
	p.tickRound()

	select {
	case <-p.jobCh:
		t.Fatal("未注册链路不应派发轮询任务")
	case <-time.After(20 * time.Millisecond):
	}
}
