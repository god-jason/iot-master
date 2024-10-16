package connect

import (
	"io"
	"sync"
	"time"
)

type Messenger struct {
	Conn

	Timeout time.Duration

	mu sync.Mutex
}

func (m *Messenger) Ask(request []byte, response []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//s := bufio.NewReader(m.Conn)

	//先写
	_, err := m.Conn.Write(request)
	if err != nil {
		return 0, err
	}

	//读超时
	err = m.Conn.SetReadTimeout(m.Timeout)
	if err != nil {
		return 0, err
	}

	return m.Conn.Read(response)
}

func (m *Messenger) AskAtLeast(request []byte, response []byte, min int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//先写
	_, err := m.Conn.Write(request)
	if err != nil {
		return 0, err
	}

	//读超时
	err = m.Conn.SetReadTimeout(m.Timeout)
	if err != nil {
		return 0, err
	}

	return io.ReadAtLeast(m.Conn, response, min)
}

func (m *Messenger) Read(response []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//读超时
	err := m.Conn.SetReadTimeout(m.Timeout)
	if err != nil {
		return 0, err
	}
	//读
	return m.Conn.Read(response)
}

func (m *Messenger) ReadAtLeast(response []byte, min int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//读超时
	err := m.Conn.SetReadTimeout(m.Timeout)
	if err != nil {
		return 0, err
	}

	return io.ReadAtLeast(m.Conn, response, min)
}
