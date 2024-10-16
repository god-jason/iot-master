package modbus

import "github.com/god-jason/bucket/log"

type Filter struct {
	Name       string `json:"name"`       //字段
	Expression string `json:"expression"` //表达式
	//Entire     bool   `json:"entire"`
}

type Calculator struct {
	Name       string `json:"name"`       //赋值
	Expression string `json:"expression"` //表达式
}

type Poller struct {
	Code    uint8  `json:"code"`
	Address uint16 `json:"address"`
	Length  uint16 `json:"length"` //长度
}

func (p *Poller) Parse(mapper *Mapper, buf []byte, values map[string]any) error {
	switch p.Code {
	case 1:
		for _, m := range mapper.Coils {
			if p.Address <= m.Address && p.Length > m.Address-p.Address {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				values[m.Name] = ret
			}
		}
	case 2:
		for _, m := range mapper.DiscreteInputs {
			if p.Address <= m.Address && p.Length > m.Address-p.Address {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				values[m.Name] = ret
			}
		}
	case 3:
		for _, m := range mapper.HoldingRegisters {
			if p.Address <= m.Address && p.Length > m.Address-p.Address {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				//03 指令 的 位类型
				if rets, ok := ret.(map[string]bool); ok {
					for k, v := range rets {
						values[k] = v
					}
				} else {
					values[m.Name] = ret
				}
			}
		}
	case 4:
		for _, m := range mapper.HoldingRegisters {
			if p.Address <= m.Address && p.Length > m.Address-p.Address {
				ret, err := m.Parse(p.Address, buf)
				if err != nil {
					log.Error(err)
					continue
				}
				//04 指令 的 位类型
				if rets, ok := ret.(map[string]bool); ok {
					for k, v := range rets {
						values[k] = v
					}
				} else {
					values[m.Name] = ret
				}
			}
		}

	}

	return nil
}
