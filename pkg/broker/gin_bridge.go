package broker

import (
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/busy-cloud/boat/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = &websocket.Upgrader{
	//HandshakeTimeout: time.Second,
	ReadBufferSize:  512,
	WriteBufferSize: 512,
	Subprotocols:    []string{"mqtt"},
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func GinBridge(ctx *gin.Context) {
	c, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer c.Close()

	//阻塞执行
	err = server.EstablishConnection("web", &wsConn{Conn: c.NetConn(), Raw: c})
	if err != nil {
		log.Error(err)
	}
	ctx.Abort()
}

// wsConn is a websocket connection which satisfies the net.Conn interface.
type wsConn struct {
	net.Conn
	Raw *websocket.Conn
	r   io.Reader
}

// Read reads the next span of bytes from the websocket connection and returns the number of bytes read.
func (ws *wsConn) Read(p []byte) (int, error) {
	if ws.r == nil {
		op, r, err := ws.Raw.NextReader()
		if err != nil {
			return 0, err
		}

		if op != websocket.BinaryMessage {
			err = errors.New("must be binary")
			return 0, err
		}

		ws.r = r
	}

	var err error
	var n, br int
	for {
		br, err = ws.r.Read(p[n:])
		n += br

		if err != nil {
			ws.r = nil
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return n, err
		}

		if n == len(p) {
			return n, err
		}
	}
}

// Write writes bytes to the websocket connection.
func (ws *wsConn) Write(p []byte) (int, error) {
	err := ws.Raw.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close signals the underlying websocket conn to close.
func (ws *wsConn) Close() error {
	return ws.Conn.Close()
}
