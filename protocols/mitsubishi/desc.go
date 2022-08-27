package mitsubishi

import (
	"iot-master/connect"
	"iot-master/protocols/protocol"
)

var Fx_Program = protocol.Desc{
	Name:    "Fx-Program",
	Version: "1.0",
	Label:   "Fx-Program",
	Factory: func(tunnel connect.Tunnel, opts protocol.Options) protocol.Protocol {
		return &FxProgram{
			link: tunnel,
		}
	},
	Parser: ParseFxProgramAddress,
}
