package vconn

import (
	"errors"
	"io"
	"net"
	"time"
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
	if e1 != nil && e2 != nil {
		return errors.Join(e1, e2)
		//return errors.New(e1.Error() + " and " + e2.Error())
	}
	if e1 != nil {
		return e1
	}
	if e2 != nil {
		return e2
	}
	return nil
}

func (c *VConn) LocalAddr() net.Addr                { return &vAddr{} }
func (c *VConn) RemoteAddr() net.Addr               { return &vAddr{} }
func (c *VConn) SetDeadline(t time.Time) error      { return nil }
func (c *VConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *VConn) SetWriteDeadline(t time.Time) error { return nil }

func New() (*VConn, *VConn) {
	var c1, c2 VConn
	c1.PipeReader, c2.PipeWriter = io.Pipe()
	c2.PipeReader, c1.PipeWriter = io.Pipe()
	return &c1, &c2
}
