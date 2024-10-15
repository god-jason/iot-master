package vconn

import "testing"

func TestVConn_data(t *testing.T) {
	c1, c2 := New()
	go func() {
		buf := make([]byte, 128)
		n, _ := c1.Read(buf)
		t.Log("c1 recv", string(buf[:n]))
		_, _ = c1.Write([]byte("world"))
	}()

	_, _ = c2.Write([]byte("hello"))

	buf := make([]byte, 128)
	n, _ := c2.Read(buf)
	t.Log("c2 recv", string(buf[:n]))

}
