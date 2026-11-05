// Package mqtt — 通过订阅外置 MQTT Broker 接入设备。
//
// 提供 MQTT 客户端连接/重连、按 $thing/{device_id}/{suffix} 的上下行主题订阅与发布。
// Device 组合 iot.BaseDevice，出站指令通过 Publish 下发并等待设备 ACK（按消息 id 匹配，带超时）。
// 外置 broker 场景下无 CONNECT/DISCONNECT hook，subscribe 收到第一帧消息时构造并 Register。
package mqtt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"

	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/pool"
)

// CustomConnectionFunc 虚拟连接函数（直通 broker、免进程间通讯）。
// 由外部注入——master 内嵌 broker 时用 bridge 函数。
var CustomConnectionFunc paho.OpenConnectionFunc

// client 全局 paho 客户端实例
var client paho.Client

// Client 返回全局客户端
func Client() paho.Client { return client }

// Options 启动配置（显式传优先生效，否则从 viper 读）
type Options struct {
	URL      string // broker 地址，如 tcp://127.0.0.1:1883；为空表示不主动连 broker（内嵌直通场景）
	ClientID string // 客户端 ID，为空时自动生成 12 位随机串
	Username string // 连接用户名
	Password string // 连接密码
}

// Startup 启动 MQTT 客户端
// - enable=false 跳过连接
// - opts 有字段时优先用 opts，否则从 viper 读
// - URL 为空时跳过 broker 连接，供 CustomConnectionFunc 内嵌场景用
func Startup(opts ...Options) error {
	if !Enabled() {
		logger.Info("mqtt 模块已禁用，跳过启动")
		return nil
	}

	// 解析配置：显式传 > viper
	o := Options{}
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.URL == "" {
		o.URL = config.GetString(MODULE, "url")
	}
	if o.ClientID == "" {
		o.ClientID = config.GetString(MODULE, "client_id")
	}
	if o.Username == "" {
		o.Username = config.GetString(MODULE, "username")
	}
	if o.Password == "" {
		o.Password = config.GetString(MODULE, "password")
	}

	pahoOpts := paho.NewClientOptions()
	// 非 Windows 平台默认追加本地 unix socket broker（进程内直通通信使用）
	if runtime.GOOS != "windows" {
		pahoOpts.AddBroker("unix:///var/run/boat.sock")
	}
	if o.URL != "" {
		pahoOpts.AddBroker(o.URL)
	}
	if o.ClientID == "" {
		o.ClientID = uuid.NewString()[:12]
	}
	pahoOpts.SetClientID(o.ClientID)
	pahoOpts.SetUsername(o.Username)
	pahoOpts.SetPassword(o.Password)
	// 断线自动重连
	pahoOpts.SetConnectRetry(true)
	// 心跳间隔 50s
	pahoOpts.SetKeepAlive(50 * 1000000000)
	// 保留会话，重连后 broker 端仍保留订阅关系
	pahoOpts.SetCleanSession(false)
	// 重连成功后恢复既有订阅
	pahoOpts.SetResumeSubs(true)
	// 注入虚拟连接：直通内嵌 broker，免进程间通讯
	if CustomConnectionFunc != nil {
		pahoOpts.SetCustomOpenConnectionFn(CustomConnectionFunc)
	}

	client = paho.NewClient(pahoOpts)
	// 连接并最多等待 10s，超时视为启动失败
	token := client.Connect()
	if !token.WaitTimeout(10 * 1000000000) {
		return fmt.Errorf("mqtt connect timeout")
	}
	return token.Error()
}

// Shutdown 断开连接
func Shutdown() error {
	if client != nil {
		client.Disconnect(250)
	}
	return nil
}

// errToken 用于在客户端未初始化或序列化失败时返回明确错误。
// Publish 的签名保持返回 paho.Token，调用方通过 token.Error() 即可感知失败。
type errToken struct{ err error }

func (t *errToken) Wait() bool { return true }

func (t *errToken) WaitTimeout(time.Duration) bool { return true }

func (t *errToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (t *errToken) Error() error { return t.err }

// Publish 发布消息（自动序列化 JSON）
// 字符串、[]byte、bytes.Buffer 原样发送，其余类型 JSON 序列化为 []byte；QoS 0、不保留。
func Publish(topic string, payload any) paho.Token {
	c := client
	if c == nil {
		return &errToken{err: fmt.Errorf("mqtt: 客户端未初始化（模块未启用或未启动）")}
	}
	switch payload.(type) {
	case string, []byte, bytes.Buffer:
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			return &errToken{err: fmt.Errorf("mqtt: 消息序列化失败: %w", err)}
		}
		payload = b
	}
	return c.Publish(topic, 0, false, payload)
}

// PublishEx 向多个 topic 发布同一条消息
func PublishEx(topics []string, payload any) error {
	c := client
	if c == nil {
		return fmt.Errorf("mqtt: 客户端未初始化（模块未启用或未启动）")
	}
	var data interface{}
	switch payload.(type) {
	case string, []byte, bytes.Buffer:
		data = payload
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("mqtt: 消息序列化失败: %w", err)
		}
		data = b
	}
	for _, topic := range topics {
		c.Publish(topic, 0, false, data)
	}
	return nil
}

// callback 消息回调签名：入参为原始 topic 与 payload
type callback func(topic string, payload []byte)

// subs 各 topic filter 已注册的回调集合；同一 filter 只向 broker 订阅一次
// subsMu 保护 subs：回调协程读取与 Subscribe 写入可能并发发生
var (
	subsMu sync.RWMutex
	subs   = make(map[string][]callback)
)

// Subscribe 订阅 topic filter，回调通过协程池并发执行
func Subscribe(filter string, cb func(topic string, payload []byte)) {
	subsMu.Lock()
	if cbs, ok := subs[filter]; ok {
		// 已订阅过该 filter：只追加回调，避免重复向 broker 订阅
		subs[filter] = append(cbs, cb)
		subsMu.Unlock()
		return
	}
	subs[filter] = []callback{cb}
	subsMu.Unlock()

	client.Subscribe(filter, 0, func(c paho.Client, msg paho.Message) {
		// 持读锁拷贝一份回调列表，释放锁后再执行，避免回调内并发读写 map
		subsMu.RLock()
		cbs := append([]callback(nil), subs[filter]...)
		subsMu.RUnlock()
		// 每个消息交给协程池并发分发给该 filter 下的全部回调
		for _, ccb := range cbs {
			pool.Submit(func() { ccb(msg.Topic(), msg.Payload()) })
		}
	})
}

// SubscribeStruct 订阅并自动反序列化为结构体
func SubscribeStruct[T any](filter string, cb func(topic string, data *T)) {
	Subscribe(filter, func(topic string, payload []byte) {
		var value T
		if len(payload) > 0 {
			if err := json.Unmarshal(payload, &value); err != nil {
				logger.Error("mqtt SubscribeStruct 反序列化失败", "error", err, "topic", topic)
				return
			}
		}
		cb(topic, &value)
	})
}
