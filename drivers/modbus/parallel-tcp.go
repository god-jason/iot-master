package modbus

import (
	"errors"
	"fmt"
	"github.com/god-jason/iot-master/connect"
	"github.com/god-jason/iot-master/pkg/bin"
	"sync"
	"time"
)

type response struct {
	buf []byte
	err error
}

type request struct {
	cmd  []byte
	resp chan response //out
}

// ParallelTCP Modbus-TCP协议
type ParallelTCP struct {
	tunnel connect.Conn
	queue  chan interface{} //in

	requests  sync.Map
	increment uint16
}

func NewParallelTCP(tunnel connect.Conn, opts string) *ParallelTCP {
	concurrency := 10

	tcp := &ParallelTCP{
		tunnel:    tunnel,
		queue:     make(chan interface{}, concurrency),
		increment: 0x0A0A, //避免首字节为0，有些DTU可能会异常
	}
	return tcp
}

func (m *ParallelTCP) execute(cmd []byte, immediate bool) ([]byte, error) {
	req := &request{
		cmd:  cmd,
		resp: make(chan response, 1),
	}

	if !immediate {
		//排队等待
		m.queue <- nil
	}

	id := m.increment
	bin.WriteUint16(cmd, id) //写入事务ID
	m.increment++
	if m.increment < 0x0A0A {
		m.increment = 0x0A0A
	}
	m.requests.Store(id, req)

	//TODO 下发指令
	_, err := m.tunnel.Write(cmd)
	if err != nil {
		if !immediate {
			//释放队列
			<-m.queue
		}
		return nil, err
	}

	//等待结果
	select {
	case <-time.After(5 * time.Second):
		if !immediate {
			<-m.queue //清空
		}
		return nil, errors.New("timeout")
	case resp := <-req.resp:
		return resp.buf, resp.err
	}
}

func (m *ParallelTCP) OnData(buf []byte) {
	//取出请求，并让出队列，可以开始下一个请示了
	if len(m.queue) > 0 {
		<-m.queue
	}

	//解析数据
	l := len(buf)
	if l < 10 {
		return //长度不够
	}

	length := bin.ParseUint16(buf[4:])
	packLen := int(length) + 6

	if l < packLen {
		//TODO 缓存包，下次再处理？？？
		return //长度不够
	}

	//处理数据包
	m.handlePack(buf[:packLen])

	//粘包处理
	//如果有剩余内容，可能是下一个响应包，需要继续处理
	//此代码会导致后包比前包先处理
	if l > packLen {
		m.OnData(buf[packLen:])
	}
}

func (m *ParallelTCP) handlePack(buf []byte) {
	id := bin.ParseUint16(buf)
	r, ok := m.requests.LoadAndDelete(id)
	if !ok {
		return
	}
	req := r.(*request)

	//slave := buf[6]
	fc := buf[7]
	//解析错误码
	if fc&0x80 > 0 {
		req.resp <- response{err: fmt.Errorf("错误码：%d", buf[2])}
		return
	}

	//解析数据
	//length := 4
	count := int(buf[8])
	switch fc {
	case 1, 2:
		//数组解压
		bb := bin.ExpandBool(buf[9:], count)
		req.resp <- response{buf: bb}
	case 3, 4, 23:
		req.resp <- response{buf: bin.Dup(buf[9:])}
	case 5, 15, 6, 16:
		//写指令不处理
	default:
		req.resp <- response{err: fmt.Errorf("错误功能码：%d", fc)}
	}
}

func (m *ParallelTCP) Read(slave uint8, code uint8, addr uint16, size uint16) ([]byte, error) {
	b := make([]byte, 12)
	//bin.WriteUint16(b, id)
	bin.WriteUint16(b[2:], 0) //协议版本
	bin.WriteUint16(b[4:], 6) //剩余长度
	b[6] = slave
	b[7] = code
	bin.WriteUint16(b[8:], addr)
	bin.WriteUint16(b[10:], uint16(size))

	return m.execute(b, true)
}

func (m *ParallelTCP) Write(slave uint8, code uint8, addr uint16, buf []byte) error {
	length := len(buf)
	//如果是线圈，需要Shrink
	switch code {
	case 1:
		if length == 1 {
			code = 5
			//数据 转成 0x0000 0xFF00
			if buf[0] > 0 {
				buf = []byte{0xFF, 0}
			} else {
				buf = []byte{0, 0}
			}
		} else {
			code = 15 //0x0F
			//数组压缩
			b := bin.ShrinkBool(buf)
			count := len(b)
			buf = make([]byte, 3+count)
			bin.WriteUint16(buf, uint16(length))
			buf[2] = uint8(count)
			copy(buf[3:], b)
		}
	case 3:
		if length == 2 {
			code = 6
		} else {
			code = 16 //0x10
			b := make([]byte, 3+length)
			bin.WriteUint16(b, uint16(length/2))
			b[2] = uint8(length)
			copy(b[3:], buf)
			buf = b
		}
	default:
		return errors.New("功能码不支持")
	}

	l := 10 + len(buf)
	b := make([]byte, l)
	//bin.WriteUint16(b, id)
	bin.WriteUint16(b[2:], 0) //协议版本
	bin.WriteUint16(b[4:], 6) //剩余长度
	b[6] = slave
	b[7] = code
	bin.WriteUint16(b[8:], addr)
	copy(b[10:], buf)

	_, err := m.execute(b, true)
	return err
}
