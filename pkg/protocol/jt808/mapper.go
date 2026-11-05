package jt808

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/thing"
)

// MapItem JT808 点位 —— 被动上报型，无主动 Poll
// 地址格式: "msgID:fieldName" 如 "0x0200:alarm"
type MapItem struct {
	Name      string
	Address   string // "msgID:fieldName" 或十进制 "512:alarm"
	DataType  string
	Scale     float64
	Offset    float64
	Precision int
	MsgID     uint16 `json:"-"`
	Field     string `json:"-"`
}

// JT808Mapper JT808 被动上报 Mapper
// 特点：不 EncodeRequest（协议不主动 Poll），只 DecodeFromReport
type JT808Mapper struct {
	items []MapItem
}

// NewMapper 从物模型创建 Mapper
func NewMapper(props []thing.Property) (*JT808Mapper, error) {
	items := make([]MapItem, 0, len(props))
	for _, p := range props {
		if p.Address == "" {
			continue
		}
		msgID, field, err := parseAddr(p.Address)
		if err != nil {
			return nil, fmt.Errorf("jt808 property %q: %w", p.Key, err)
		}
		items = append(items, MapItem{
			Name:      p.Key,
			Address:   p.Address,
			DataType:  firstNonEmpty(p.Format, "uint16"),
			Scale:     p.Rate,
			Offset:    p.Correct,
			Precision: p.Precision,
			MsgID:     msgID,
			Field:     field,
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("jt808: 无可映射属性")
	}
	return &JT808Mapper{items: items}, nil
}

func (m *JT808Mapper) Items() []MapItem { return m.items }

// DecodeFromReport 从已解析的 msgID + 字段 map 里取值
func (m *JT808Mapper) DecodeFromReport(msgID uint16, fields map[string]any) map[string]any {
	result := make(map[string]any)
	for _, it := range m.items {
		if it.MsgID != msgID {
			continue
		}
		if raw, ok := fields[it.Field]; ok {
			fv := toFloat(raw)*it.Scale + it.Offset
			result[it.Name] = applyPrecision(fv, it.Precision)
		}
	}
	return result
}

func parseAddr(s string) (uint16, string, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("JT808 地址应为 msgID:fieldName 格式")
	}
	msgID, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 0, 16)
	if err != nil {
		return 0, "", fmt.Errorf("msgID 解析失败: %w", err)
	}
	return uint16(msgID), strings.TrimSpace(parts[1]), nil
}

func toFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case uint16:
		return float64(t)
	case uint32:
		return float64(t)
	default:
		return 0
	}
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
