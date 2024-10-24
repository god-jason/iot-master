package serial

import (
	"errors"
	"github.com/god-jason/bucket/log"
	"github.com/god-jason/iot-master/connect/tunnel"
	"github.com/god-jason/iot-master/db"
	"go.bug.st/serial"
	"time"
)

func init() {
	db.Register(new(Serial))
}

// Serial 串口
type Serial struct {
	tunnel.Tunnel `xorm:"extends"`

	PortName   string `json:"port_name,omitempty"`   //port, e.g. COM1 "/dev/ttySerial1".
	BaudRate   uint   `json:"baud_rate,omitempty"`   //9600 115200
	DataBits   uint   `json:"data_bits,omitempty"`   //5 6 7 8
	StopBits   uint   `json:"stop_bits,omitempty"`   //1 2
	ParityMode int    `json:"parity_mode,omitempty"` //0 1 2 NONE ODD EVEN
}

// Open 打开
func (s *Serial) Open() error {
	if s.Running {
		return errors.New("serial is opened")
	}
	s.Closed = false

	//守护
	s.Keep(s.Open)

	opts := serial.Mode{
		BaudRate: int(s.BaudRate),
		DataBits: int(s.DataBits),
		StopBits: serial.StopBits(s.StopBits),
		Parity:   serial.Parity(s.ParityMode),
	}

	log.Trace("create serial ", s.PortName, opts)
	port, err := serial.Open(s.PortName, &opts)
	if err != nil {
		return err
	}
	s.Running = true
	s.Status = "正常"

	//读超时
	err = port.SetReadTimeout(time.Second * 5)
	if err != nil {
		return err
	}

	s.Conn = port

	return s.Start()
}
