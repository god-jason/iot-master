package tunnel

import (
	"errors"
	"github.com/god-jason/iot-master/connect"
	"github.com/god-jason/iot-master/device"
	"github.com/god-jason/iot-master/driver"
	"github.com/god-jason/iot-master/log"
	"github.com/god-jason/iot-master/pool"
	"github.com/god-jason/iot-master/types"
	"go.bug.st/serial"
	"io"
	"net"
	"time"
)

type PollerOptions struct {
	PollerPeriod   uint `json:"poller_period,omitempty"`   //采集周期
	PollerInterval uint `json:"poller_interval,omitempty"` //采集间隔
}

type ProtocolOptions struct {
	ProtocolName    string         `json:"protocol_name,omitempty"`    //协议 rtu tcp parallel-tcp
	ProtocolOptions map[string]any `json:"protocol_options,omitempty"` //协议参数
}

type RetryOptions struct {
	RetryTimeout uint `json:"retry_timeout,omitempty"` //重试时间
	RetryMaximum uint `json:"retry_maximum,omitempty"` //最大次数
}

type Tunnel struct {
	Id          string `json:"id,omitempty" xorm:"pk"` //ID
	Name        string `json:"name,omitempty"`         //名称
	Description string `json:"description,omitempty"`  //说明
	Heartbeat   string `json:"heartbeat,omitempty"`    //心跳包

	//协议
	ProtocolName    string        `json:"protocol_name,omitempty"`
	ProtocolOptions types.Options `json:"protocol_options,omitempty"`

	Disabled bool      `json:"disabled"`
	Created  time.Time `json:"created" xorm:"created"` //创建时间

	Status string `json:"status,omitempty" xorm:"-"` //状态

	Running bool `json:"-" xorm:"-"`
	Closed  bool `json:"-" xorm:"-"`

	Conn connect.Conn `json:"-" xorm:"-"`

	Adapter driver.Adapter `json:"-" xorm:"-"`

	//设备
	devices []*device.Device `json:"-" xorm:"-"`

	//透传
	pipe io.ReadWriteCloser

	//保持
	keeping bool
}

func (l *Tunnel) ID() string {
	return l.Id
}

func (l *Tunnel) Keep(open func() error) {
	if l.keeping {
		return
	}

	l.keeping = true
	go func() {
		for {
			//10秒自动重连
			time.Sleep(time.Second * 10)

			if l.Running {
				continue
			}
			if l.Closed {
				break
			}

			err := open()
			if err != nil {
				log.Error(err)
			}
		}
		l.keeping = false
	}()
}

// Close 关闭
func (l *Tunnel) Close() error {
	if l.Closed {
		return errors.New("tunnel is closed")
	}

	l.Running = false
	l.Closed = true
	l.Status = "关闭"

	if l.pipe != nil {
		_ = l.pipe.Close()
	}

	return l.Conn.Close()
}

// Write 写
func (l *Tunnel) Write(data []byte) (int, error) {
	if !l.Running {
		return 0, errors.New("tunnel closed")
	}
	if l.pipe != nil {
		return 0, nil //透传模式下，直接抛弃
	}
	//log.Trace(l.Id, "write", data)
	n, err := l.Conn.Write(data)
	if err != nil {
		//关闭连接
		_ = l.Conn.Close()
		l.Running = false
		l.Status = "关闭"
	}
	return n, err
}

// Read 读
func (l *Tunnel) Read(data []byte) (int, error) {
	if !l.Running {
		return 0, errors.New("tunnel closed")
	}
	if l.pipe != nil {
		//先read，然后透传
		return 0, nil //透传模式下，直接抛弃
	}
	//log.Trace(l.Id, "read")
	n, err := l.Conn.Read(data)
	if err != nil {
		//网络错误（读超时除外）
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return 0, err
		}

		//串口错误（读超时除外）
		var se *serial.PortError
		if errors.As(err, &se) && (se.Code() == serial.InvalidTimeoutValue) {
			return 0, err
		}

		//其他错误，关闭连接
		_ = l.Conn.Close()
		l.Running = false
		l.Status = "关闭"
	} else if n == 0 {
		//关闭连接（已知 串口会进入假死）
		_ = l.Conn.Close()
		l.Running = false
		l.Status = "关闭"
		return 0, errors.New("没有读取到数据，但是也没有报错，关掉再试")
	}
	//log.Trace(l.Id, "readed", data[:n])
	return n, err
}

func (l *Tunnel) SetReadTimeout(t time.Duration) error {
	return l.Conn.SetReadTimeout(t)
}

func (l *Tunnel) Pipe(pipe io.ReadWriteCloser) {
	//关闭之前的透传
	if l.pipe != nil {
		_ = l.pipe.Close()
	}

	l.pipe = pipe

	//传入空，则关闭
	if pipe == nil {
		return
	}

	buf := make([]byte, 1024)
	for {
		n, err := pipe.Read(buf)
		if err != nil {
			//if err == io.EOF {
			//	continue
			//}
			//pipe关闭，则不再透传
			break
		}
		//将收到的数据转发出去
		n, err = l.Conn.Write(buf[:n])
		if err != nil {
			//发送失败，说明连接失效
			_ = pipe.Close()
			break
		}
	}
	l.pipe = nil

	//TODO 使用io.copy
	//go io.Copy(pipe, l.conn)
	//go io.Copy(l.conn, pipe)
}

func (l *Tunnel) Start() (err error) {
	//加载协议
	l.Adapter, err = driver.Create(l.Conn, l.ProtocolName, l.ProtocolOptions)
	if err != nil {
		return err
	}

	l.devices, err = device.LoadByTunnel(l.Id)
	if err != nil {
		return err
	}

	//加载设备
	for _, d := range l.devices {
		d.SetAdapter(l.Adapter)

		err = l.Adapter.Mount(d.Id, d.ProductId, d.Station)
		if err != nil {
			log.Error(err)
			//return err
		}
	}

	go l.Poll()

	return nil
}

func (l *Tunnel) Poll() {

	//设备上线
	//!!! 不能这样做，不然启动服务器会产生大量的消息
	//for _, dev := range adapter.index {
	//	topic := fmt.Sprintf("device/online/%s", dev.Id)
	//	_ = mqtt.Publish(topic, nil)
	//}

	interval := l.ProtocolOptions.Int64("poller_interval", 60) //默认1分钟轮询一次
	if interval < 1 {
		interval = 1
	}

	//按毫秒计时
	interval *= 1000

	//OUT:
	for {
		start := time.Now().UnixMilli()
		for _, dev := range l.devices {
			values, err := l.Adapter.Poll(dev.Id)
			if err != nil {
				log.Error(err)
				continue
			}

			//d := device.Get(dev.Id)
			if values != nil && len(values) > 0 {
				_ = pool.Insert(func() {
					dev.Push(values)
				})

				//TODO 过滤器

				//TODO 计算器
			}
		}

		//检查连接，避免空等待
		if !l.Running {
			break
		}

		//轮询间隔
		now := time.Now().UnixMilli()
		elapsed := now - start
		if elapsed < interval {
			time.Sleep(time.Millisecond * time.Duration(interval-elapsed))
		}

		//避免空转，睡眠1分钟（延迟10ms太长，睡1分钟也有点长）
		if elapsed < 10 {
			time.Sleep(time.Minute)
		}
	}

	log.Info("modbus adapter quit", l.Id)

	//设备下线
	//for _, dev := range adapter.devices {
	//	topic := fmt.Sprintf("device/%s/offline", dev.Id)
	//	_ = mqtt.Publish(topic, nil)
	//}

	//TODO d.SetAdapter(nil)
}
