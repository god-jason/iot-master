package modbus

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/thing"
)

// ============================================================
// ModbusModel —— 一个协议物模型
//
//   - Items 是解析后的点位（FC/StartAddr/Quantity 已填好）
//   - Pollers 是批量读请求（显式指定，否则从 Items 自动合并生成），Read/Poll 共用
//   - FindItemByName / DecodeResponse 提供上层调用接口
//
// 创建路径：NewModel(id, proto, properties, pollers)
// ============================================================

// ModbusModel 协议物模型（纯解码使用，不含后处理）
//
//   - Items 是解析后的点位（FC/StartAddr/Quantity 已填好）
//   - Pollers 是批量读请求（显式指定，否则从 Items 自动合并生成）
//   - FindItemByName / DecodeResponse 提供上层调用接口
type ModbusModel struct {
	Id    string // modelId（缓存 key）
	Proto string // TCP 或 RTU
	Name  string

	Items   []MapItem      // 解析后的点位
	Pollers []ModbusPoller // 批量读请求

	itemByName map[string]*MapItem
	itemsByFC  map[byte][]*MapItem // FC → 该 FC 下的点位（DecodeResponse 按 FC 取候选）
}

// NewModel 从 thing.Property + thing.Poller 创建 ModbusModel
func NewModel(id, proto string, properties []thing.Property, pollers []thing.Poller) (*ModbusModel, error) {
	m := &ModbusModel{
		Id:   id,
		Proto: proto,
		Name: fmt.Sprintf("modbus-%s", proto),
	}

	if len(properties) == 0 {
		return m, nil
	}

	items := thingPropsToMapItems(properties)
	for i := range items {
		if err := parseModbusItem(&items[i]); err != nil {
			return nil, fmt.Errorf("item[%d] %q: %w", i, items[i].Name, err)
		}
	}
	m.Items = items

	m.buildIndex()

	if len(pollers) > 0 {
		for _, p := range pollers {
			mp, err := newModbusPoller(proto, p)
			if err != nil {
				return nil, fmt.Errorf("poller %+v: %w", p, err)
			}
			m.Pollers = append(m.Pollers, *mp)
		}
	} else {
		// 未显式指定 poller 时，从 items 自动生成（按 FC + 相邻地址合并）
		m.Pollers = append(m.Pollers, autoGeneratePollers(items)...)
	}

	return m, nil
}

// buildIndex 构建 itemByName（按 key）与 itemsByFC（按 FC）两份索引。
// itemsByFC 让 DecodeResponse 只需遍历同 FC 的点位，避免每次响应全表扫描。
func (m *ModbusModel) buildIndex() {
	m.itemByName = make(map[string]*MapItem, len(m.Items))
	m.itemsByFC = make(map[byte][]*MapItem)
	for i := range m.Items {
		it := &m.Items[i]
		m.itemByName[it.Name] = it
		m.itemsByFC[it.FC] = append(m.itemsByFC[it.FC], it)
	}
}

// itemsForFC 返回该 FC 下的点位；索引未构建时退化为全量（结果仍然正确）
func (m *ModbusModel) itemsForFC(fc byte) []*MapItem {
	if m.itemsByFC == nil {
		out := make([]*MapItem, len(m.Items))
		for i := range m.Items {
			out[i] = &m.Items[i]
		}
		return out
	}
	return m.itemsByFC[fc]
}

// FindItemByName 按 key 查找
func (m *ModbusModel) FindItemByName(name string) *MapItem {
	if m == nil || m.itemByName == nil {
		return nil
	}
	return m.itemByName[name]
}

// findPollerFor 找到包含 target 的批量读 poller
func (m *ModbusModel) findPollerFor(target *MapItem) *ModbusPoller {
	for i := range m.Pollers {
		p := &m.Pollers[i]
		if p.fc == target.FC &&
			target.StartAddr >= p.startAddr &&
			target.StartAddr+target.Quantity <= p.startAddr+p.quantity {
			return p
		}
	}
	return nil
}

// DecodeResponse 把响应 Frame 转成语义值
func (m *ModbusModel) DecodeResponse(slaveAddr, unitID byte, resp *Frame) (map[string]any, error) {
	if resp == nil {
		return nil, fmt.Errorf("modbus 响应错误: resp is nil")
	}
	if resp.IsError {
		return nil, fmt.Errorf("modbus 响应错误: %s", resp.ErrorMsg)
	}

	result := make(map[string]any)
	values := resp.Uint16s()
	switch resp.Function {
	case FCReadCoils, FCReadDiscreteInputs:
		for _, it := range m.itemsForFC(resp.Function) {
			if it.FC != resp.Function {
				continue
			}
			if it.StartAddr >= resp.StartAddr+resp.Quantity || it.StartAddr < resp.StartAddr {
				continue
			}
			idx := it.StartAddr - resp.StartAddr
			if int(idx) >= len(values) {
				continue
			}
			semantic := applyBitResult(values[idx] == 1, it.Scale, it.Offset, it.Precision)
			if len(it.Enums) > 0 {
				if mapped, ok := mapToEnum(semantic, it.Enums); ok {
					semantic = mapped
				}
			}
			result[it.Name] = semantic
		}

	case FCReadHoldingRegisters, FCReadInputRegisters:
		// Data 是响应 PDU（不含 FC），首字节 byteCount 之后即寄存器数据
		var respBytes []byte
		if len(resp.Data) > 1 {
			respBytes = resp.Data[1:]
		}

		for _, it := range m.itemsForFC(resp.Function) {
			if it.FC != resp.Function {
				continue
			}
			if it.StartAddr+it.Quantity > resp.StartAddr+resp.Quantity || it.StartAddr < resp.StartAddr {
				continue
			}
			off := it.StartAddr - resp.StartAddr
			byteOff := int(off) * 2
			byteLen := int(it.Quantity) * 2
			if byteOff+byteLen > len(respBytes) {
				continue
			}
			rawBytes := respBytes[byteOff : byteOff+byteLen]
			lo, sw := it.LittleEnd, it.Swap

			var semantic any
			if it.Bit > 0 {
				if len(rawBytes) >= 2 {
					word := uint16(rawBytes[0])<<8 | uint16(rawBytes[1])
					bitVal := (word>>(it.Bit-1))&1 == 1
					semantic = applyBitResult(bitVal, it.Scale, it.Offset, it.Precision)
				} else {
					semantic = applyBitResult(false, it.Scale, it.Offset, it.Precision)
				}
			} else {
				val, err := Decode(rawBytes, it.DataType, lo, sw)
				if err != nil {
					return result, err
				}
				semantic = applyScaleOffsetPrecision(val, it.Scale, it.Offset, it.Precision)
			}

			if len(it.Enums) > 0 {
				if mapped, ok := mapToEnum(semantic, it.Enums); ok {
					semantic = mapped
				}
			}
			result[it.Name] = semantic
		}

	default:
		return result, fmt.Errorf("modbus FC=%02X 不是读响应", resp.Function)
	}

	return result, nil
}

// buildWritePDU 构造写请求（返回不含 FC 的 pdu + FC）。
// TCP/RTU 共用，帧编码由 transport.encode 统一拼 FC。
func buildWritePDU(target *MapItem, value any) (byte, []byte, error) {
	enumVal, err := reverseEnum(value, target.Enums)
	if err != nil {
		return 0, nil, err
	}
	scale := target.Scale
	if scale == 0 {
		scale = 1
	}
	rawFloat := (float64(enumVal) - target.Offset) / scale

	lo, sw := target.LittleEnd, target.Swap
	payload, err := Encode(rawFloat, target.DataType, lo, sw)
	if err != nil {
		return 0, nil, err
	}

	isCoil := target.FC == FCReadCoils || target.FC == FCReadDiscreteInputs

	var fc byte
	var pdu []byte
	if isCoil {
		fc = FCWriteSingleCoil
		valWord := uint16(0x0000)
		if rawFloat != 0 {
			valWord = 0xFF00
		}
		pdu = []byte{byte(target.StartAddr >> 8), byte(target.StartAddr), byte(valWord >> 8), byte(valWord)}
	} else if target.Quantity > 1 {
		fc = FCWriteMultipleRegs
		byteCount := len(payload)
		pdu = make([]byte, 0, 5+byteCount)
		pdu = append(pdu, byte(target.StartAddr>>8), byte(target.StartAddr),
			byte(byteCount/2>>8), byte(byteCount/2), byte(byteCount))
		pdu = append(pdu, payload...)
	} else {
		fc = FCWriteSingleRegister
		var valWord uint16
		if len(payload) >= 2 {
			valWord = uint16(payload[0])<<8 | uint16(payload[1])
		}
		pdu = []byte{byte(target.StartAddr >> 8), byte(target.StartAddr), byte(valWord >> 8), byte(valWord)}
	}
	return fc, pdu, nil
}

// ============================================================
// ModbusDevice —— 挂载的子设备
// ============================================================

// ModbusDevice 挂载的子设备
type ModbusDevice struct {
	Id      string // deviceId（gateway 层给的）
	Name    string // 可选显示名
	Model   string // 绑定的 modelId（从 models cache 按 Id 反查）
	Address string // 原始地址字符串（如 "1" / "0x01"）
	Slave   byte   // 解析后的从站号
}

// NewDevice 创建 ModbusDevice（立即解析 address → Slave）
func NewDevice(id, name, modelId, address string) (*ModbusDevice, error) {
	slave, ok := parseAddrOK(address)
	if !ok {
		return nil, fmt.Errorf("modbus device %q address %q 无法解析", id, address)
	}
	return &ModbusDevice{
		Id:      id,
		Name:    name,
		Model:   modelId,
		Address: address,
		Slave:   slave,
	}, nil
}

// ============================================================
// thing → MapItem 辅助 + parseAddr
// ============================================================

func thingPropsToMapItems(props []thing.Property) []MapItem {
	items := make([]MapItem, 0, len(props))
	for _, p := range props {
		enums := make([]EnumDef, 0, len(p.Enums))
		for _, e := range p.Enums {
			enums = append(enums, EnumDef{Key: e.Key, Name: e.Name, Value: e.Value})
		}
		items = append(items, MapItem{
			Name:      p.Key,
			Address:   p.Address,
			DataType:  parseDataType(p.Format),
			LittleEnd: p.LittleEndian,
			Scale:     p.Rate,
			Offset:    p.Correct,
			Precision: p.Precision,
			Bit:       p.Bit,
			Enums:     enums,
		})
	}
	return items
}

// parseModbusItem 往 MapItem 里填 FC/StartAddr/Quantity
func parseModbusItem(item *MapItem) error {
	if item.Scale == 0 {
		item.Scale = 1
	}

	spec, err := ParseAddr(item.Address, item.DataType, 0)
	if err != nil {
		return err
	}
	item.FC = spec.FC
	item.StartAddr = spec.StartAddr
	item.Quantity = spec.Quantity
	return nil
}

// parseAddrOK 把地址字符串解析成 byte（从站号）
// 返回 ok=false 表示解析失败
func parseAddrOK(s string) (byte, bool) {
	s = strings.TrimSpace(s)
	if v, err := strconv.ParseUint(s, 10, 8); err == nil {
		return byte(v), true
	}
	if v, err := strconv.ParseUint(s, 16, 8); err == nil {
		return byte(v), true
	}
	return 0, false
}
