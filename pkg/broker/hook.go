package broker

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"net"
	"time"

	"github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

type Hook struct {
	mqtt.HookBase
	Key string
}

func (h *Hook) ID() string {
	return "broker"
}
func (h *Hook) Provides(b byte) bool {
	//高效吗？
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
		mqtt.OnDisconnect,
		mqtt.OnSubscribed,
		mqtt.OnUnsubscribed,
	}, []byte{b})
}

func (h *Hook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	if conn, ok := cl.Net.Conn.(*net.TCPConn); ok {
		//_ = conn.SetKeepAlive(false) //避免服务器主动下发rst，导致设备无法低功耗
		_ = conn.SetKeepAlivePeriod(10 * time.Minute) //10分钟，慢了点
	}

	_ = Publish("client/"+cl.ID+"/connect", nil)
	return nil
}

func (h *Hook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	//log.Info("MQTT incoming ", cl.Net.Listener, " ", cl.Net.Remote, " ", cl.ID)

	//if cl.Net.Listener == "web" {
	//	return true
	//}

	switch cl.Net.Listener {
	case "internal":
		return true
	//unix websocket 直接鉴权通过
	case "unix", "web":
		return true
	case "base":
		//log.Info("[base] OnConnectAuthenticate ", cl.ID, pk.Connect.Username, pk.Connect.Password)

		//不支持匿名
		if !pk.Connect.UsernameFlag {
			return false
		}

		//检测用户名 和 密码
		pas := md5.Sum([]byte(string(pk.Connect.Username) + h.Key))
		if string(pk.Connect.Password) != hex.EncodeToString(pas[:]) {
			return false
		}

		return true
	default:
		return false
	}
}

func (h *Hook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	//只允许发送属性事件

	return true
}

func (h *Hook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	//执行unsubscribe
	_ = Publish("client/"+cl.ID+"/disconnect", nil)

	//server.Clients.Get(cl.ID)

}

func (h *Hook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	//_ = Publish("client/"+cl.ID+"/subscribe", pk.Filters)
}

func (h *Hook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	//_ = Publish("client/"+cl.ID+"/unsubscribe", pk.Payload)
}
