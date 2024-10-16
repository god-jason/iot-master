package modbus

type Mapper struct {
	Coils            []*PointBit  `json:"coils,omitempty"`
	DiscreteInputs   []*PointBit  `json:"discrete_inputs,omitempty"`
	HoldingRegisters []*PointWord `json:"holding_registers,omitempty"`
	InputRegisters   []*PointWord `json:"input_registers,omitempty"`
}

func (p *Mapper) Lookup(name string) (pt Point, code uint8, address uint16) {
	for _, m := range p.Coils {
		if m.Name == name {
			return m, 1, m.Address
		}
	}

	for _, m := range p.DiscreteInputs {
		if m.Name == name {
			return m, 2, m.Address
		}
	}

	for _, m := range p.HoldingRegisters {
		if m.Name == name {
			return m, 3, m.Address
		}
	}

	for _, m := range p.InputRegisters {
		if m.Name == name {
			return m, 4, m.Address
		}
	}
	return nil, 0, 0
}
