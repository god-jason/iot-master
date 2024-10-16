package dlt645

import (
	"fmt"
)

type Mapper struct {
	Points []*Point `json:"Point,omitempty"`
}

func (p *Mapper) Lookup(name string) (addr string, err error) {
	for _, m := range p.Points {
		if m.Name == name {
			return m.Id, nil
		}
	}

	return "", fmt.Errorf("找不到点位 %s", name)
}
