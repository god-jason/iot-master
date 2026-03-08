package lib

import (
	"io"
	"net"
	"time"

	"go.uber.org/multierr"
)

type vAddr struct {
}

func (a *vAddr) Network() string {
	return "virtual"
}

func (a *vAddr) String() string {
	return "virtual"
}

type VConn struct {
	*io.PipeReader
	*io.PipeWriter
}

func (c *VConn) Close() error {
	e1 := c.PipeWriter.Close()
	e2 := c.PipeReader.Close()
	return multierr.Append(e1, e2)
}

func (c *VConn) LocalAddr() net.Addr                { return &vAddr{} }
func (c *VConn) RemoteAddr() net.Addr               { return &vAddr{} }
func (c *VConn) SetDeadline(t time.Time) error      { return nil }
func (c *VConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *VConn) SetWriteDeadline(t time.Time) error { return nil }

func NewVConn() (*VConn, *VConn) {
	var c1, c2 VConn
	c1.PipeReader, c2.PipeWriter = io.Pipe()
	c2.PipeReader, c1.PipeWriter = io.Pipe()
	return &c1, &c2
}
