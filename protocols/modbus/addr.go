package modbus

import (
	"errors"
	"github.com/zgwit/iot-master/protocol"
	"regexp"
	"strconv"
)

type Address struct {
	Code   uint8  `json:"code"`
	Offset uint16 `json:"offset"`
}

func (a *Address) String() string {

	return ""
}

var addrRegexp *regexp.Regexp

func init() {
	addrRegexp, _ = regexp.Compile(`^(C|D|DI|H|I)(\d+)$`)

}

func ParseAddress(add string) (protocol.Addr, error) {
	ss := addrRegexp.FindStringSubmatch(add)
	if ss == nil || len(ss) != 3 {
		return nil, errors.New("unknown address")
	}
	var code uint8 = 1
	switch ss[1] {
	case "C":
		code = 1
	case "D":
		fallthrough
	case "DI":
		code = 2
	case "H":
		code = 3
	case "I":
		code = 4
	}
	offset, _ := strconv.ParseUint(ss[2], 10, 10)
	//offset, _ := strconv.Atoi(ss[2])
	return &Address{
		Code:   code,
		Offset: uint16(offset),
	}, nil
}

// TODO const
var DescRTU = protocol.Describer{
	Name:    "ModbusRTU",
	Version: "1.0",
	Label:   "Modbus RTU",
	Factory: newRTU,
	Address: ParseAddress,
}

var DescTCP = protocol.Describer{
	Name:    "ModbusTCP",
	Version: "1.0",
	Label:   "Modbus TCP",
	Factory: newTCP,
	Address: ParseAddress,
}