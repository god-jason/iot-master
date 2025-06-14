package protocol

import "github.com/busy-cloud/boat/lib"

var protocols lib.Map[Protocol]

func GetProtocols() []*Base {
	var b []*Base
	protocols.Range(func(name string, p *Protocol) bool {
		b = append(b, &p.Base)
		return true
	})
	return b
}

func GetProtocol(name string) *Protocol {
	return protocols.Load(name)
}

func Register(protocol *Protocol) {
	protocols.Store(protocol.Name, protocol)
}
