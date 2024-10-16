package server

import (
	"errors"
	"fmt"
	"github.com/god-jason/bucket/log"
	"github.com/zgwit/iot-master/connect"
	"github.com/zgwit/iot-master/connect/tunnel"
	"github.com/zgwit/iot-master/db"
	"github.com/zgwit/iot-master/driver"
	"net"
	"regexp"
)

var snRegex *regexp.Regexp

func init() {
	db.Register(new(Server))

	//限定字符串
	snRegex = regexp.MustCompile("^\\w+$")
}

// Server TCP服务器
type Server struct {
	tunnel.Tunnel `xorm:"extends"`

	Port uint16 `json:"port,omitempty"` //监听端口
	//Standalone bool   `json:"standalone,omitempty"` //单例模式（不支持注册）

	Mode string `json:"mode,omitempty"` //normal普通，standalone单例，register注册包

	children map[string]*Link

	listener *net.TCPListener
}

func (s *Server) handleSingle(c *net.TCPConn) (err error) {
	const k = "internal"
	if cc, ok := s.children[k]; ok {
		_ = cc.Close()
	}

	l := &Link{
		Tunnel:   s.Tunnel, //TODO 删除lock
		ServerId: s.Id,
		Remote:   c.RemoteAddr().String(),
	}
	l.Running = true
	l.Status = "正常"
	l.Conn = &connect.NetConn{Conn: c}

	s.children[k] = l
	//以ServerID保存
	links.Store(s.Id, l)

	return l.Start()
}

func (s *Server) handleIncoming(c *net.TCPConn) error {
	var err error
	l := &Link{
		Tunnel:   s.Tunnel,
		ServerId: s.Id,
		Remote:   c.RemoteAddr().String(),
	}

	l.Running = true
	l.Status = "正常"
	l.Conn = &connect.NetConn{Conn: c}

	s.Adapter, err = driver.Create(l, s.ProtocolName, s.ProtocolOptions)
	return err
}

func (s *Server) handleRegister(c *net.TCPConn) error {
	//TODO 只有配置了注册包，才能正常通讯
	buf := make([]byte, 128)
	n, err := c.Read(buf)
	if err != nil {
		_ = c.Close()
		return err
	}

	data := buf[:n]
	//检查，必须是合法的字符串，字母、数字、汉字、下划线
	//ret, err := regexp.MatchString("^\\w+$", sn)
	if !snRegex.Match(data) {
		_, _ = c.Write([]byte("invalid sn"))
		_ = c.Close()
		return err
	}

	sn := string(data)

	var l Link
	//get, err := db.Engine.Where("server_id=?", s.Id).And("sn=?", sn).Get(&conn)
	get, err := db.Engine.ID(sn).Get(&l)
	if err != nil {
		_, _ = c.Write([]byte("database error"))
		_ = c.Close()
		return err
	}

	if !get {
		l = Link{
			Tunnel:   s.Tunnel,
			ServerId: s.Id,
			Remote:   c.RemoteAddr().String(),
		}
		l.Id = sn //修改ID

		_, err = db.Engine.InsertOne(&l)
		if err != nil {
			_, _ = c.Write([]byte("database error"))
			_ = c.Close()
			return err
		}
	}

	l.Running = true
	l.Status = "正常"
	l.Conn = &connect.NetConn{c}

	s.children[sn] = &l
	links.Store(l.Id, &l)

	return l.Start()
}

// Open 打开
func (s *Server) Open() error {
	if s.Running {
		return errors.New("server is opened")
	}
	s.Closed = false

	log.Trace("create server ", s.Port)
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		s.Status = err.Error()
		return err
	}

	//守护
	s.Keep(s.Open)

	s.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		s.Status = err.Error()
		return err
	}
	s.Running = true
	s.Status = "正常"

	s.children = make(map[string]*Link)
	go func() {
		for {
			c, err := s.listener.AcceptTCP()
			if err != nil {
				//TODO 需要正确处理接收错误
				log.Error(err)
				s.Status = err.Error()
				break
			}

			switch s.Mode {
			case "register":
				err = s.handleRegister(c)
			case "single":
				err = s.handleSingle(c)
			case "multiple", "":
				err = s.handleIncoming(c)
			default:
				//应该无法启动
				_ = c.Close()
			}

			if err != nil {
				log.Error(err)
				s.Status = err.Error()
			}
		}

		s.Running = false
	}()

	return nil
}

// Close 关闭
func (s *Server) Close() error {
	s.Running = false
	s.Closed = true

	//close tunnels
	if s.children != nil {
		for _, l := range s.children {
			_ = l.Close()
		}
	}

	return s.listener.Close()
}

// GetTunnel 获取连接
func (s *Server) GetTunnel(id string) *Link {
	return s.children[id]
}
