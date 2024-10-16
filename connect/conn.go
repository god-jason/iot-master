package connect

import (
	"io"
	"time"
)

type Conn interface {
	io.ReadWriteCloser
	SetReadTimeout(t time.Duration) error
}
