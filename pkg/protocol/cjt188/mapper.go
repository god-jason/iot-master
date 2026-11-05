package cjt188

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/protocol/meter"
	"github.com/god-jason/iot-master/pkg/thing"
)

// MapItem CJT188 点位
type MapItem struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`    // DI 格式: "00.01" / "1" / "0x0001"（2 字节）
	DataType  string  `json:"data_type"`  // meter.lua type string
	Scale     float64 `json:"scale"`
	Offset    float64 `json:"offset"`
	Precision int     `json:"precision"`
	Reverse   bool    `json:"reverse,omitempty"`
	HasUnit   bool    `json:"has_unit,omitempty"`
	UnitRev   bool    `json:"unit_reverse,omitempty"`

	Ctrl byte    `json:"-"` // 请求控制字
	DI   [2]byte `json:"-"` // 低位在前
	Size int     `json:"-"` // 值的字节长度
}

// Cjt188Mapper CJT188 Mapper —— 每个点单独发一次 read 请求
type Cjt188Mapper struct {
	items []MapItem
}

// Cjt188Request 一次请求的关键字段
type Cjt188Request struct {
	Ctrl    byte
	DI      [2]byte
	Payload []byte
}

// NewMapper 从 thing.Property 列表创建 Mapper
func NewMapper(props []thing.Property) (*Cjt188Mapper, error) {
	items := make([]MapItem, 0, len(props))
	for _, p := range props {
		if p.Address == "" {
			continue
		}
		di, err := parseDI(p.Address)
		if err != nil {
			return nil, fmt.Errorf("cjt188 property %q address %q: %w", p.Key, p.Address, err)
		}
		dt := firstNonEmpty(p.Format, "XXXX")
		size, _ := meter.FormatSize(dt)
		items = append(items, MapItem{
			Name:      p.Key,
			Address:   p.Address,
			DataType:  dt,
			Scale:     p.Rate,
			Offset:    p.Correct,
			Precision: p.Precision,
			Reverse:   p.LittleEndian,
			Ctrl:      CtrlReadData,
			DI:        di,
			Size:      size,
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("cjt188: 无可映射属性")
	}
	return &Cjt188Mapper{items: items}, nil
}

func (m *Cjt188Mapper) Items() []MapItem { return m.items }

// FindItemByName 按 key 查找点位
func (m *Cjt188Mapper) FindItemByName(name string) *MapItem {
	for i := range m.items {
		if m.items[i].Name == name {
			return &m.items[i]
		}
	}
	return nil
}

// EncodeRequests 为每个点位生成一次独立请求
func (m *Cjt188Mapper) EncodeRequests() []Cjt188Request {
	out := make([]Cjt188Request, 0, len(m.items))
	for _, it := range m.items {
		out = append(out, Cjt188Request{Ctrl: it.Ctrl, DI: it.DI})
	}
	return out
}

// DecodeResponse 用原始请求 DI + 响应 payload 解码
func (m *Cjt188Mapper) DecodeResponse(reqDI [2]byte, respData []byte) (map[string]any, error) {
	result := make(map[string]any)

	var matched *MapItem
	for i := range m.items {
		if m.items[i].DI[0] == reqDI[0] && m.items[i].DI[1] == reqDI[1] {
			matched = &m.items[i]
			break
		}
	}
	if matched == nil {
		return result, fmt.Errorf("cjt188: DI [%02X %02X] 未注册", reqDI[0], reqDI[1])
	}

	data := respData
	var unitByte byte
	var hasUnitByte bool

	if matched.HasUnit {
		if matched.UnitRev {
			if len(data) < 1 {
				return result, fmt.Errorf("cjt188: 单位前置但响应空")
			}
			unitByte = data[0]
			data = data[1:]
			hasUnitByte = true
		} else {
			if len(data) > matched.Size {
				unitByte = data[matched.Size]
				data = data[:matched.Size]
				hasUnitByte = true
			}
		}
	}

	semantic, err := meter.Decode(data, matched.DataType, matched.Reverse, matched.Scale, matched.Offset, matched.Precision)
	if err != nil {
		return result, fmt.Errorf("cjt188 item %q: %w", matched.Name, err)
	}

	if hasUnitByte {
		if f, ok := semantic.(float64); ok {
			semantic = meter.UnitConvert(unitByte, f)
		}
	}

	result[matched.Name] = semantic
	return result, nil
}

// parseDI 解析 DI
func parseDI(addr string) ([2]byte, error) {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return [2]byte{}, fmt.Errorf("DI 为空")
	}

	if strings.Contains(trimmed, ".") {
		parts := strings.Split(trimmed, ".")
		if len(parts) > 2 {
			return [2]byte{}, fmt.Errorf("DI 点分段数 > 2")
		}
		var hi, lo byte
		if len(parts) == 1 {
			v, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 0, 16)
			if err != nil {
				return [2]byte{}, fmt.Errorf("DI %q: %w", parts[0], err)
			}
			lo = byte(v & 0xFF)
			hi = byte(v >> 8)
		} else {
			v1, err1 := strconv.ParseUint(strings.TrimSpace(parts[0]), 0, 8)
			v2, err2 := strconv.ParseUint(strings.TrimSpace(parts[1]), 0, 8)
			if err1 != nil || err2 != nil {
				return [2]byte{}, fmt.Errorf("DI 解析失败")
			}
			hi = byte(v1)
			lo = byte(v2)
		}
		return [2]byte{lo, hi}, nil
	}

	v, err := strconv.ParseUint(trimmed, 0, 16)
	if err != nil {
		return [2]byte{}, fmt.Errorf("DI %q 无法解析: %w", trimmed, err)
	}
	return [2]byte{byte(v & 0xFF), byte(v >> 8)}, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// lint: thing 包确保被引用
var _ = thing.Poller{}
