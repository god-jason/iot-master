package dlt645

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/protocol/meter"
	"github.com/god-jason/iot-master/pkg/thing"
)

// MapItem DLT645 点位
type MapItem struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"` // DI 格式: "90100100" / "0x90100100" / "90.10.01.00" / "XXXXXX.XX" (meter.lua type)
	DataType  string  `json:"data_type"` // meter.lua type string，如 "XXXXXX.XX" / "uint16" / "HHHH"
	Scale     float64 `json:"scale"`
	Offset    float64 `json:"offset"`
	Precision int     `json:"precision"`
	Reverse   bool    `json:"reverse,omitempty"` // 是否逆序（多字节值）

	DI   uint32 `json:"-"` // 解析后的 DI
	Size int    `json:"-"` // 值的字节长度
}

// DLT645Mapper DLT645 Mapper —— 每个点单独读一次 DI
type DLT645Mapper struct {
	items []MapItem
}

// NewMapper 从物模型创建 Mapper
func NewMapper(props []thing.Property) (*DLT645Mapper, error) {
	items := make([]MapItem, 0, len(props))
	for _, p := range props {
		if p.Address == "" {
			continue
		}
		di, err := parseDI(p.Address)
		if err != nil {
			return nil, fmt.Errorf("dlt645 property %q: %w", p.Key, err)
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
			DI:        di,
			Size:      size,
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("dlt645: 无可映射属性")
	}
	return &DLT645Mapper{items: items}, nil
}

func (m *DLT645Mapper) Items() []MapItem { return m.items }

// FindItemByName 按 key 查找点位
func (m *DLT645Mapper) FindItemByName(name string) *MapItem {
	for i := range m.items {
		if m.items[i].Name == name {
			return &m.items[i]
		}
	}
	return nil
}

// EncodeRequest 生成每个 item 独立的 read 请求 DI（4 字节 DI，低位在前）
func (m *DLT645Mapper) EncodeRequest() [][]byte {
	out := make([][]byte, 0, len(m.items))
	for _, it := range m.items {
		di := it.DI
		out = append(out, []byte{
			byte(di & 0xFF),
			byte((di >> 8) & 0xFF),
			byte((di >> 16) & 0xFF),
			byte((di >> 24) & 0xFF),
		})
	}
	return out
}

// DecodeResponse 从响应 Data（已 sub33）里解码
// 响应 Data 格式：[DI 4B 低位在前] + [value bytes...]
func (m *DLT645Mapper) DecodeResponse(respCtrl byte, respData []byte) (map[string]any, error) {
	result := make(map[string]any)
	if len(respData) < 4 {
		return result, fmt.Errorf("dlt645: 响应长度不足 need>=4 have=%d", len(respData))
	}
	di := uint32(respData[0]) | uint32(respData[1])<<8 | uint32(respData[2])<<16 | uint32(respData[3])<<24
	valBytes := respData[4:]

	var matched *MapItem
	for i := range m.items {
		if m.items[i].DI == di {
			matched = &m.items[i]
			break
		}
	}
	if matched == nil {
		return result, fmt.Errorf("dlt645: DI 0x%08X 未注册", di)
	}

	semantic, err := meter.Decode(valBytes, matched.DataType, matched.Reverse, matched.Scale, matched.Offset, matched.Precision)
	if err != nil {
		return result, fmt.Errorf("dlt645 item %q: %w", matched.Name, err)
	}
	result[matched.Name] = semantic

	_ = respCtrl
	return result, nil
}

// parseDI 把地址字符串解析成 uint32（4 字节 DI）
func parseDI(addr string) (uint32, error) {
	s := strings.TrimSpace(addr)
	if s == "" {
		return 0, fmt.Errorf("DI 为空")
	}
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if len(parts) > 4 {
			return 0, fmt.Errorf("DI 点分段数 > 4")
		}
		var out uint32
		shift := uint((len(parts) - 1) * 8) // 第一段在最高位
		for _, p := range parts {
			// 每段按 16 进制解析（DI 段 "90" 应该是 0x90）
			v, err := strconv.ParseUint(strings.TrimSpace(p), 16, 8)
			if err != nil {
				return 0, fmt.Errorf("DI 段 %q: %w", p, err)
			}
			out |= uint32(v) << shift
			shift -= 8
		}
		return out, nil
	}
	// 去掉 0x 前缀
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	// 裸数字：先尝试按 16 进制（DLT645 DI 通常是 hex 格式），失败再十进制
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		if v2, err2 := strconv.ParseUint(s, 10, 32); err2 == nil {
			v = v2
		} else {
			return 0, fmt.Errorf("DI 解析失败: %w", err)
		}
	}
	if v > 0xFFFFFFFF {
		return 0, fmt.Errorf("DI 超出 uint32")
	}
	return uint32(v), nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// lint: sort imported but not used
var _ = sort.Slice
