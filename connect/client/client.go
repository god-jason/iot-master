package client

import (
	"errors"
	"fmt"
	"github.com/god-jason/bucket/log"
	"github.com/zgwit/iot-master/connect"
	"github.com/zgwit/iot-master/connect/tunnel"
	"github.com/zgwit/iot-master/db"
	"net"
)

func init() {
	db.Register(new(Client))
}

// Client 网络链接
type Client struct {
	tunnel.Tunnel `xorm:"extends"`

	Net  string `json:"net,omitempty"`  //类型 tcp udp
	Addr string `json:"addr,omitempty"` //地址，主机名或IP
	Port uint16 `json:"port,omitempty"` //端口号
}

// Open 打开
func (c *Client) Open() error {
	if c.Running {
		return errors.New("client is opened")
	}
	c.Closed = true

	//守护
	c.Keep(c.Open)

	//发起连接
	addr := fmt.Sprintf("%s:%d", c.Addr, c.Port)
	log.Trace("create client ", addr)
	conn, err := net.Dial(c.Net, addr)
	if err != nil {
		//time.AfterFunc(time.Minute, client.RetryOptions)
		c.Status = err.Error()
		return err
	}
	c.Running = true
	c.Status = "正常"

	c.Conn = &connect.NetConn{Conn: conn}

	return c.Start()
}
