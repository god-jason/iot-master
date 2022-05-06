package connect

import (
	"fmt"
	"github.com/zgwit/iot-master/database"
	"github.com/zgwit/iot-master/model"
)

//NewTunnel 创建通道
func NewTunnel(tunnel *model.Tunnel) (Tunnel, error) {
	var tnl Tunnel
	switch tunnel.Type {
	case "tcp-client":
		tnl = newNetClient(tunnel, "tcp")
		break
	case "tcp-server":
		tnl = newTcpServer(tunnel)
		break
	case "udp-client":
		tnl = newNetClient(tunnel, "udp")
		break
	case "udp-server":
		tnl = newUdpServer(tunnel)
		break
	case "serial":
		tnl = newSerial(tunnel)
		break
	default:
		return nil, fmt.Errorf("Unsupport type %s ", tunnel.Type)
	}

	tnl.On("open", func() {
		_ = database.History.Save(model.Event{
			Target: "tunnel",
			TargetId: tunnel.Id,
			Event:    "打开",
		})
	})

	tnl.On("close", func() {
		_ = database.History.Save(model.Event{
			Target: "tunnel",
			TargetId: tunnel.Id,
			Event:    "关闭",
		})
	})

	tnl.On("link", func(conn Link) {
		_ = database.History.Save(model.Event{
			Target: "link",
			TargetId: conn.Id(),
			Event:  "上线",
		})
		conn.Once("close", func() {
			_ = database.History.Save(model.Event{
				Target: "link",
				TargetId: conn.Id(),
				Event:  "离线",
			})
		})
	})

	return tnl, nil
}
