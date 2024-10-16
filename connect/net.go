package connect

import (
	"net"
	"time"
)

type NetConn struct {
	net.Conn
}

func (c *NetConn) SetReadTimeout(t time.Duration) error {
	return c.SetReadDeadline(time.Now().Add(t))
}
