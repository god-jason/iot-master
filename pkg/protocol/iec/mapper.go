package iec

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/thing"
)

// MapItem IEC 104 点位
// 地址格式: "type:IOA" 如 "1:1234" — type 是 ASDU type (1=M_SP, 3=M_ME_NC, 30=M_ME_NC_NA)
type MapItem struct {
	Name      string
	Address   string // "type:IOA"
	DataType  string
	Scale     float64
	Offset    float64
	Precision int
	ASDUType  byte   `json:"-"`
	IOA       uint32 `json:"-"`
}

// IECMapper IEC 104 Mapper
type IECMapper struct {
	items []MapItem
}

// NewMapper 从物模型创建 Mapper
func NewMapper(props []thing.Property) (*IECMapper, error) {
	items := make([]MapItem, 0, len(props))
	for _, p := range props {
		if p.Address == "" {
			continue
		}
		t, ioa, err := parseAddr(p.Address)
		if err != nil {
			return nil, fmt.Errorf("iec property %q: %w", p.Key, err)
		}
		items = append(items, MapItem{
			Name:      p.Key,
			Address:   p.Address,
			DataType:  firstNonEmpty(p.Format, "float32"),
			Scale:     p.Rate,
			Offset:    p.Correct,
			Precision: p.Precision,
			ASDUType:  t,
			IOA:       ioa,
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("iec: 无可映射属性")
	}
	return &IECMapper{items: items}, nil
}

func (m *IECMapper) Items() []MapItem { return m.items }

// DecodeFromASDU 从原始 ASDU 解析值
func (m *IECMapper) DecodeFromASDU(asduType byte, ioas map[uint32]float64) map[string]any {
	result := make(map[string]any)
	for _, it := range m.items {
		if it.ASDUType != asduType {
			continue
		}
		if raw, ok := ioas[it.IOA]; ok {
			fv := raw*it.Scale + it.Offset
			result[it.Name] = applyPrecision(fv, it.Precision)
		}
	}
	return result
}

func parseAddr(s string) (byte, uint32, error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("IEC 地址应为 type:IOA 格式")
	}
	t, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 0, 8)
	if err != nil {
		return 0, 0, fmt.Errorf("type 解析失败: %w", err)
	}
	ioa, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 0, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("IOA 解析失败: %w", err)
	}
	return byte(t), uint32(ioa), nil
}

func applyPrecision(v float64, p int) float64 {
	if p <= 0 {
		return v
	}
	pow := 1.0
	for i := 0; i < p; i++ {
		pow *= 10
	}
	return float64(int64(v*pow+0.5)) / pow
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
