package protocol

import (
	"github.com/busy-cloud/boat/mqtt"
)

const RegisterTopic = "iot/protocol/register"

func Register(p *Protocol) {
	mqtt.Publish("iot/protocol/register", p)
	//protocols.Store(protocol.Name, protocol)
}
