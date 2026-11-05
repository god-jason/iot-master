// mqtttest MQTT 两种接入方式的全 Topic 测试程序。
//
// 用法：
//
//	go run ./cmd/mqtttest -mode=embedded
//	go run ./cmd/mqtttest -mode=external -addr=tcp://127.0.0.1:1883
//
// 两种模式都使用同一个模拟设备，覆盖 docs/MQTT.md 中定义的全部上下行 Topic：
//
//	上行：property/report、event/report、status/report（及对应 response）
//	下行：property/set、property/get、service/call、setting/set、
//	      setting/get、database/sync、database/load（及对应 response）
//
// 模式说明：
//   - embedded：启动进程内嵌 broker（internal/broker + pkg/beeq），模拟设备直连测试；
//   - external：连接一个真实外置 broker（EMQX/mosquitto 等），平台侧使用 internal/mqtt 客户端接入测试。
//
// 设备/物模型测试数据会写入 DB（-dsn 或 master.yaml 的 db 段）。
package main

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"xorm.io/builder"

	"github.com/god-jason/iot-master/internal/broker"
	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/mqtt"
	"github.com/god-jason/iot-master/internal/types"
	beeq "github.com/god-jason/iot-master/pkg/beeq"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/thing"
)

var (
	mode     = flag.String("mode", "embedded", "接入方式：embedded（内嵌 broker）| external（外置 broker）")
	addr     = flag.String("addr", "", "embedded 监听地址（默认 127.0.0.1:18831）；external 必填外置 broker 地址（如 tcp://127.0.0.1:1883）")
	driver   = flag.String("driver", "mysql", "DB 驱动（-dsn 生效时使用）")
	dsn      = flag.String("dsn", "", "DB DSN；为空时从 master.yaml 的 db 段拼装")
	modelID  = flag.String("model", "mqtttest-model", "测试物模型 ID")
	deviceID = flag.String("device", "mqtttest-device", "测试设备 ID")
	waitSec  = flag.Int("wait", 5, "单条消息等待超时（秒）")
)

func main() {
	flag.Parse()

	logger.Init("info")

	drv, link, err := resolveDSN(*dsn, *driver)
	if err != nil {
		exitf("解析 DB 连接失败: %v", err)
	}
	fx, err := prepareFixtures(drv, link, *modelID, *deviceID)
	if err != nil {
		exitf("准备测试数据失败: %v", err)
	}

	switch *mode {
	case "embedded":
		runEmbedded(fx)
	case "external":
		runExternal(fx)
	default:
		exitf("未知 -mode=%s（可选 embedded / external）", *mode)
	}
}

// ============ 测试数据 ============

type fixtures struct {
	modelID  string
	deviceID string
	secret   string
}

func prepareFixtures(drv, link, mid, did string) (fixtures, error) {
	if err := db.InitEngine(drv, link); err != nil {
		return fixtures{}, err
	}

	model := types.Model{
		ID:     mid,
		Name:   "MQTT 测试物模型",
		Secret: db.NewSecret(),
		Model: thing.Model{
			Protocol: "mqtt",
			Properties: []thing.Property{
				{Key: "temperature", Name: "温度", Type: "float", Mode: "rw"},
				{Key: "humidity", Name: "湿度", Type: "float", Mode: "rw"},
			},
		},
	}

	var existing types.Model
	found, err := db.FindOne("model", builder.Eq{"id": mid}, &existing)
	if err != nil {
		return fixtures{}, fmt.Errorf("查询物模型: %w", err)
	}
	if found {
		model.Secret = existing.Secret
		if model.Secret == "" {
			model.Secret = db.NewSecret()
			if _, err := db.UpdateOne("model", builder.Eq{"id": mid}, map[string]interface{}{"secret": model.Secret}); err != nil {
				return fixtures{}, err
			}
		}
	} else if _, err := db.InsertOne("model", &model); err != nil {
		return fixtures{}, fmt.Errorf("创建物模型: %w", err)
	}

	var dev types.Device
	found, err = db.FindOne("device", builder.Eq{"id": did}, &dev)
	if err != nil {
		return fixtures{}, fmt.Errorf("查询设备: %w", err)
	}
	if found {
		if dev.ModelID != mid || dev.Disabled {
			if _, err := db.UpdateOne("device", builder.Eq{"id": did}, map[string]interface{}{
				"model_id": mid,
				"disabled": false,
			}); err != nil {
				return fixtures{}, err
			}
		}
	} else {
		dev = types.Device{ID: did, SN: "mqtttest-" + did, Name: "MQTT 测试设备", ModelID: mid}
		if _, err := db.InsertOne("device", &dev); err != nil {
			return fixtures{}, fmt.Errorf("创建设备: %w", err)
		}
	}

	return fixtures{modelID: mid, deviceID: did, secret: model.Secret}, nil
}

func resolveDSN(flagDSN, flagDriver string) (string, string, error) {
	if flagDSN != "" {
		return flagDriver, flagDSN, nil
	}

	config.Name("master")
	config.AddPath(".")
	_ = config.Load(false)

	// 优先使用分散字段（master.yaml 的 db.driver/host/...）
	if drv := config.GetString("db", "driver"); drv != "" {
		link, err := db.DSN(
			drv,
			config.GetString("db", "host"),
			config.GetInt("db", "port"),
			config.GetString("db", "database"),
			config.GetString("db", "username"),
			config.GetString("db", "password"),
			config.GetString("db", "sslmode"),
			config.GetString("db", "params"),
		)
		return drv, link, err
	}

	if u := config.GetString("db", "url"); u != "" {
		if t := config.GetString("db", "type"); t != "" {
			return t, u, nil
		}
		return "mysql", u, nil
	}

	drv := config.GetString("db", "driver")
	if drv == "" {
		drv = "mysql"
	}
	link, err := db.DSN(
		drv,
		config.GetString("db", "host"),
		config.GetInt("db", "port"),
		config.GetString("db", "database"),
		config.GetString("db", "username"),
		config.GetString("db", "password"),
		config.GetString("db", "sslmode"),
		config.GetString("db", "params"),
	)
	return drv, link, err
}

// ============ embedded：内嵌 broker 接入 ============

func runEmbedded(fx fixtures) {
	listen := *addr
	if listen == "" {
		listen = "127.0.0.1:18831"
	}

	svc := broker.NewService()
	server := beeq.NewServer(beeq.DefaultOptions(), svc)
	svc.Attach(server)

	go func() {
		if err := server.Listen("tcp://" + listen); err != nil {
			logger.Error("内嵌 broker 退出", "error", err)
		}
	}()
	if err := waitTCP(listen, 3*time.Second); err != nil {
		exitf("内嵌 broker 未启动: %v", err)
	}
	defer server.Close()

	fmt.Printf("[embedded] 内嵌 broker 已监听 %s\n", listen)

	client, responder := startDeviceClient("tcp://"+listen, fx)
	if client == nil {
		os.Exit(1)
	}
	defer client.Disconnect(200)

	dev, ok := iot.GetDevice(fx.deviceID)
	if !ok {
		exitf("内嵌 broker 鉴权后设备未注册到 iot")
	}

	set := newResultSet()
	testUplinkWithAck(set, client, responder, fx)
	testDownlinkCommon(set, dev)
	testDownlinkDatabase(set, dev)
	responder.assertAllTopics(set)
	set.summary("embedded")
}

// ============ external：外置 broker 接入 ============

func runExternal(fx fixtures) {
	if *addr == "" {
		exitf("-mode=external 需要 -addr 指向一个真实 MQTT broker（容器示例：docker run -p 1883:1883 emqx/emqx）")
	}

	// 平台侧：internal/mqtt 客户端接入外置 broker
	config.Set(mqtt.MODULE, "enable", true)
	if err := mqtt.Startup(mqtt.Options{
		URL:      *addr,
		ClientID: "mqtttest-platform-" + uuid.NewString()[:8],
	}); err != nil {
		exitf("平台侧 MQTT 客户端连接失败: %v", err)
	}
	defer mqtt.Shutdown()
	mqtt.StartSubscribe()
	time.Sleep(300 * time.Millisecond)

	client, responder := startDeviceClient(*addr, fx)
	if client == nil {
		os.Exit(1)
	}
	defer client.Disconnect(200)

	set := newResultSet()
	testUplinkExternal(set, client, fx)
	dev := platformDevice(set, fx)
	testDownlinkCommon(set, dev)
	testDownlinkDatabaseByPublish(set, fx, responder)
	responder.assertAllTopics(set)
	set.summary("external")
}

// platformDevice 等待平台侧收到上行后完成设备注册，返回 iot 中的设备
func platformDevice(set *resultSet, fx fixtures) iot.Device {
	deadline := time.Now().Add(time.Duration(*waitSec) * time.Second)
	for time.Now().Before(deadline) {
		if dev, ok := iot.GetDevice(fx.deviceID); ok {
			return dev
		}
		time.Sleep(50 * time.Millisecond)
	}
	set.check("iot 设备注册", fmt.Errorf("平台侧未在 %ds 内注册设备 %s", *waitSec, fx.deviceID))
	return nil
}

// ============ 上行 ============

func uplinkCases() []struct {
	name    string
	suffix  string
	id      string
	payload any
} {
	propID := uuid.NewString()
	eventID := uuid.NewString()
	statusID := uuid.NewString()

	event := mqtt.EventReportMsg{ID: eventID}
	event.Event.Name = "alarm"
	event.Event.Type = "alert"
	event.Event.Params = map[string]interface{}{"alarmType": "over_temperature", "value": 31.2}

	return []struct {
		name    string
		suffix  string
		id      string
		payload any
	}{
		{"property/report", mqtt.PropertyReport, propID, mqtt.PropertyReportMsg{
			ID:     propID,
			Params: map[string]interface{}{"temperature": 25.6, "humidity": 61.5},
			Time:   time.Now().UnixMilli(),
		}},
		{"event/report", mqtt.EventReport, eventID, event},
		{"status/report", mqtt.StatusReport, statusID, mqtt.StatusReportMsg{
			ID:     statusID,
			Params: map[string]interface{}{"firmware": "v1.0.0", "rssi": -60},
			Time:   time.Now().UnixMilli(),
		}},
	}
}

func testUplinkWithAck(set *resultSet, client paho.Client, r *responder, fx fixtures) {
	for _, c := range uplinkCases() {
		if !publishJSON(client, mqtt.BuildTopic(fx.deviceID, c.suffix), c.payload) {
			set.check("上行 "+c.name, fmt.Errorf("发布失败"))
			continue
		}
		ack, err := r.waitAck(c.id, time.Duration(*waitSec)*time.Second)
		if err != nil {
			set.check("上行 "+c.name, err)
			continue
		}
		if ack.Ret != 1 {
			set.check("上行 "+c.name, fmt.Errorf("ack ret=%d message=%s", ack.Ret, ack.Message))
			continue
		}
		set.check("上行 "+c.name+" + 应答", nil)
	}
}

// testUplinkExternal 外置 broker 模式：平台无 ACK，改为校验平台侧状态落库
func testUplinkExternal(set *resultSet, client paho.Client, fx fixtures) {
	cases := uplinkCases()

	if !publishJSON(client, mqtt.BuildTopic(fx.deviceID, cases[0].suffix), cases[0].payload) {
		set.check("上行 property/report", fmt.Errorf("发布失败"))
		return
	}
	dev := platformDevice(set, fx)
	if dev == nil {
		return
	}

	deadline := time.Now().Add(time.Duration(*waitSec) * time.Second)
	for {
		if _, ok := dev.Property("temperature"); ok {
			set.check("上行 property/report", nil)
			fmt.Printf("      平台已收到 temperature=%v humidity=%v\n", dev.Properties()["temperature"], dev.Properties()["humidity"])
			break
		}
		if time.Now().After(deadline) {
			set.check("上行 property/report", fmt.Errorf("平台属性未更新"))
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if publishJSON(client, mqtt.BuildTopic(fx.deviceID, cases[1].suffix), cases[1].payload) {
		deadline = time.Now().Add(time.Duration(*waitSec) * time.Second)
		for {
			name, _ := dev.LastEvent()
			if name == "alarm" {
				set.check("上行 event/report", nil)
				break
			}
			if time.Now().After(deadline) {
				set.check("上行 event/report", fmt.Errorf("平台未记录事件"))
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	} else {
		set.check("上行 event/report", fmt.Errorf("发布失败"))
	}

	set.check("上行 status/report", publishErr(client, mqtt.BuildTopic(fx.deviceID, cases[2].suffix), cases[2].payload))
}

// ============ 下行 ============

func testDownlinkCommon(set *resultSet, dev iot.Device) {
	if dev == nil {
		return
	}

	set.check("下行 property/set", dev.SetProperty(map[string]interface{}{"temperature": 26.5}))
	if v, err := dev.GetProperty([]string{"temperature"}); err != nil {
		set.check("下行 property/get", err)
	} else {
		set.check("下行 property/get", nil)
		fmt.Printf("      设备返回 temperature=%v\n", v["temperature"])
	}
	if data, err := dev.Control("reboot", map[string]interface{}{"delay": 5}); err != nil {
		set.check("下行 service/call", err)
	} else {
		set.check("下行 service/call", nil)
		fmt.Printf("      设备返回 data=%v\n", data)
	}
	set.check("下行 setting/set", dev.SetConfig(map[string]interface{}{"report_interval": 10}))
	if cfg, err := dev.GetConfig(nil); err != nil {
		set.check("下行 setting/get", err)
	} else {
		set.check("下行 setting/get", nil)
		fmt.Printf("      设备返回 config=%v\n", cfg)
	}
}

func testDownlinkDatabase(set *resultSet, dev iot.Device) {
	bd, ok := dev.(*broker.Device)
	if !ok {
		set.check("下行 database/sync", fmt.Errorf("设备类型不支持 database 指令: %T", dev))
		return
	}
	records := []map[string]interface{}{
		{"id": "1", "name": "设备1"},
		{"id": "2", "name": "设备2"},
	}
	if resp, err := bd.SyncDatabase("device", records); err != nil {
		set.check("下行 database/sync", err)
	} else {
		set.check("下行 database/sync", nil)
		fmt.Printf("      设备返回 data=%v\n", resp.Data)
	}
	if resp, err := bd.LoadDatabase("device"); err != nil {
		set.check("下行 database/load", err)
	} else {
		set.check("下行 database/load", nil)
		fmt.Printf("      设备返回 data=%v\n", resp.Data)
	}
}

// testDownlinkDatabaseByPublish 外置 broker 模式下 internal/mqtt.Device 未提供 database 指令，
// 直接下发 Topic 并等待设备侧确认收到（broker 转发是异步的，不能发布后立即断言）。
func testDownlinkDatabaseByPublish(set *resultSet, fx fixtures, r *responder) {
	syncMsg := mqtt.DatabaseSyncMsg{ID: uuid.NewString(), Params: mqtt.DatabaseSyncParams{
		Table: "device",
		Records: []map[string]interface{}{
			{"id": "1", "name": "设备1"},
			{"id": "2", "name": "设备2"},
		},
	}}
	loadMsg := mqtt.DatabaseLoadMsg{ID: uuid.NewString(), Params: mqtt.DatabaseLoadParams{Table: "device"}}

	set.check("下行 database/sync", mqtt.Publish(mqtt.BuildTopic(fx.deviceID, mqtt.DatabaseSync), syncMsg).Error())
	set.check("下行 database/load", mqtt.Publish(mqtt.BuildTopic(fx.deviceID, mqtt.DatabaseLoad), loadMsg).Error())

	r.waitSeen(mqtt.DatabaseSync, mqtt.DatabaseLoad)
}

// waitSeen 等待设备侧收到指定下行 Topic（最多 wait 秒）
func (r *responder) waitSeen(suffixes ...string) {
	deadline := time.Now().Add(time.Duration(*waitSec) * time.Second)
	for time.Now().Before(deadline) {
		r.mu.Lock()
		all := true
		for _, s := range suffixes {
			if !r.seen[s] {
				all = false
				break
			}
		}
		r.mu.Unlock()
		if all {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ============ 模拟设备 ============

var downlinkSuffixes = []string{
	mqtt.PropertySet, mqtt.PropertyGet, mqtt.ServiceCall,
	mqtt.SettingSet, mqtt.SettingGet, mqtt.DatabaseSync, mqtt.DatabaseLoad,
}

var ackSuffixes = []string{
	mqtt.PropertyReportResp, mqtt.EventReportResp, mqtt.StatusReportResp,
}

type responder struct {
	mu   sync.Mutex
	dev  string
	cli  paho.Client
	seen map[string]bool
	acks chan *mqtt.ReportResponseMessage

	temperature float64
	humidity    float64
	settings    map[string]interface{}
	database    map[string][]map[string]interface{}
}

func startDeviceClient(url string, fx fixtures) (paho.Client, *responder) {
	r := &responder{
		dev:         fx.deviceID,
		seen:        make(map[string]bool),
		acks:        make(chan *mqtt.ReportResponseMessage, 16),
		temperature: 25.0,
		humidity:    60.0,
		settings:    map[string]interface{}{"report_interval": 30, "log_level": "info"},
		database:    make(map[string][]map[string]interface{}),
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(url)
	opts.SetClientID(fx.deviceID)
	opts.SetUsername(fx.modelID)
	opts.SetPassword(modelPassword(fx))
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(false)

	c := paho.NewClient(opts)
	token := c.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		fmt.Printf("[device] 连接 broker 超时\n")
		return nil, r
	}
	if err := token.Error(); err != nil {
		fmt.Printf("[device] 连接 broker 失败: %v\n", err)
		return nil, r
	}
	r.cli = c

	subs := make(map[string]byte)
	for _, suffix := range downlinkSuffixes {
		subs[mqtt.BuildTopic(fx.deviceID, suffix)] = 1
	}
	for _, suffix := range ackSuffixes {
		subs[mqtt.BuildTopic(fx.deviceID, suffix)] = 1
	}
	st := c.SubscribeMultiple(subs, r.handleMessage)
	if !st.WaitTimeout(5*time.Second) || st.Error() != nil {
		fmt.Printf("[device] 订阅下行 Topic 失败: %v\n", st.Error())
		c.Disconnect(100)
		return nil, r
	}

	fmt.Printf("[device] 已连接 %s，已订阅 %d 个 Topic\n", url, len(subs))
	return c, r
}

func modelPassword(fx fixtures) string {
	content := fmt.Sprintf("%s/%s/%s", fx.modelID, fx.deviceID, fx.secret)
	h := hmac.New(md5.New, []byte(fx.secret))
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

func (r *responder) handleMessage(_ paho.Client, msg paho.Message) {
	topic := msg.Topic()
	suffix := strings.TrimPrefix(topic, mqtt.BuildTopic(r.dev, ""))

	r.mu.Lock()
	r.seen[suffix] = true
	r.mu.Unlock()

	switch suffix {
	case mqtt.PropertySet:
		var m mqtt.PropertySetMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.mu.Lock()
		if v, ok := asFloat(m.Params["temperature"]); ok {
			r.temperature = v
		}
		if v, ok := asFloat(m.Params["humidity"]); ok {
			r.humidity = v
		}
		r.mu.Unlock()
		r.reply(mqtt.PropertySetResp, mqtt.ResponseMessage{ID: m.ID, Ret: 1, Message: "success"})

	case mqtt.PropertyGet:
		var m mqtt.PropertyGetMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.mu.Lock()
		data := map[string]interface{}{"temperature": r.temperature, "humidity": r.humidity}
		r.mu.Unlock()
		r.reply(mqtt.PropertyGetResp, mqtt.ResponseMessage{ID: m.ID, Ret: 1, Message: "success", Data: data})

	case mqtt.ServiceCall:
		var m mqtt.ServiceCallMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.reply(mqtt.ServiceCallResp, mqtt.ResponseMessage{
			ID: m.ID, Ret: 1, Message: "service " + m.Service + " done",
			Data: map[string]interface{}{"executed": true},
		})

	case mqtt.SettingSet:
		var m mqtt.SettingSetMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.mu.Lock()
		for k, v := range m.Setting {
			r.settings[k] = v
		}
		r.mu.Unlock()
		r.reply(mqtt.SettingSetResp, mqtt.ResponseMessage{
			ID: m.ID, Ret: 1, Message: "config applied",
			Data: map[string]interface{}{"effective_time": time.Now().UnixMilli()},
		})

	case mqtt.SettingGet:
		var m mqtt.SettingGetMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.mu.Lock()
		data := make(map[string]interface{})
		if len(m.Keys) == 0 {
			for k, v := range r.settings {
				data[k] = v
			}
		} else {
			for _, k := range m.Keys {
				if v, ok := r.settings[k]; ok {
					data[k] = v
				}
			}
		}
		r.mu.Unlock()
		r.reply(mqtt.SettingGetResp, mqtt.ResponseMessage{ID: m.ID, Ret: 1, Message: "success", Data: data})

	case mqtt.DatabaseSync:
		var m mqtt.DatabaseSyncMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.mu.Lock()
		r.database[m.Params.Table] = m.Params.Records
		r.mu.Unlock()
		r.reply(mqtt.DatabaseSyncResp, mqtt.ResponseMessage{
			ID: m.ID, Ret: 1, Message: "success",
			Data: map[string]interface{}{"table": m.Params.Table, "count": len(m.Params.Records)},
		})

	case mqtt.DatabaseLoad:
		var m mqtt.DatabaseLoadMsg
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		r.mu.Lock()
		records := r.database[m.Params.Table]
		r.mu.Unlock()
		if records == nil {
			records = []map[string]interface{}{}
		}
		r.reply(mqtt.DatabaseLoadResp, mqtt.ResponseMessage{
			ID: m.ID, Ret: 1, Message: "success",
			Data: map[string]interface{}{"table": m.Params.Table, "records": records},
		})

	case mqtt.PropertyReportResp, mqtt.EventReportResp, mqtt.StatusReportResp:
		var m mqtt.ReportResponseMessage
		if json.Unmarshal(msg.Payload(), &m) != nil {
			return
		}
		select {
		case r.acks <- &m:
		default:
		}
	}
}

func (r *responder) reply(suffix string, payload mqtt.ResponseMessage) {
	if r.cli == nil {
		return
	}
	body, _ := json.Marshal(payload)
	token := r.cli.Publish(mqtt.BuildTopic(r.dev, suffix), 1, false, body)
	token.WaitTimeout(2 * time.Second)
}

func (r *responder) waitAck(id string, timeout time.Duration) (*mqtt.ReportResponseMessage, error) {
	deadline := time.After(timeout)
	for {
		select {
		case ack := <-r.acks:
			if ack.ID == id {
				return ack, nil
			}
		case <-deadline:
			return nil, fmt.Errorf("等待应答超时（id=%s）", id)
		}
	}
}

// assertAllTopics 校验设备侧确实收到了每一个下行 Topic
func (r *responder) assertAllTopics(set *resultSet) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, suffix := range downlinkSuffixes {
		if !r.seen[suffix] {
			set.check("Topic 覆盖 "+suffix, fmt.Errorf("设备未收到该下行 Topic"))
			continue
		}
		set.check("Topic 覆盖 "+suffix, nil)
	}
}

// ============ 工具 ============

type resultSet struct {
	mu     sync.Mutex
	passed int
	failed int
}

func newResultSet() *resultSet { return &resultSet{} }

func (s *resultSet) check(name string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.failed++
		fmt.Printf("  x %-28s %v\n", name, err)
		return
	}
	s.passed++
	fmt.Printf("  + %-28s OK\n", name)
}

func (s *resultSet) summary(mode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("\n[%s] 通过 %d 项，失败 %d 项\n", mode, s.passed, s.failed)
	if s.failed > 0 {
		os.Exit(1)
	}
	fmt.Println("全部 Topic 测试通过")
}

func publishJSON(c paho.Client, topic string, payload any) bool {
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("序列化失败: %v\n", err)
		return false
	}
	token := c.Publish(topic, 1, false, body)
	if !token.WaitTimeout(3 * time.Second) {
		fmt.Printf("发布 %s 超时\n", topic)
		return false
	}
	if err := token.Error(); err != nil {
		fmt.Printf("发布 %s 失败: %v\n", topic, err)
		return false
	}
	return true
}

func publishErr(c paho.Client, topic string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := c.Publish(topic, 1, false, body)
	if !token.WaitTimeout(3 * time.Second) {
		return fmt.Errorf("发布超时")
	}
	return token.Error()
}

func asFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	}
	return 0, false
}

func waitTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return nil
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("等待 %s 就绪超时: %v", addr, lastErr)
}

func exitf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
