package server

import (
	"errors"
	"github.com/zgwit/iot-master/db"
	"github.com/zgwit/iot-master/tunnel"
)

func init() {
	db.Register(new(Link))
}

// Link 网络连接
type Link struct {
	tunnel.Tunnel `xorm:"extends"`

	ServerId string `json:"server_id" xorm:"index"` //服务器ID
	Remote   string `json:"remote,omitempty"`       //远程地址
}

func (l *Link) Open() error {

	return errors.New("link cannot open")
}
