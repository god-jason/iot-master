package modbus

import (
	"fmt"
	"sort"

	"github.com/god-jason/iot-master/pkg/thing"
)

// pollerSep 相邻点位合并为一次批量读的间隔阈值：地址差 ≤ sep 视为连续区间。
const pollerSep = 3

// ModbusPoller 显式轮询器 —— 表示一次批量读请求（FC + StartAddr + Quantity）
type ModbusPoller struct {
	fc        byte   // 功能码
	startAddr uint16 // 起始地址（0-based）
	quantity  uint16 // 读取数量（coil 为 bit 数，register 为 16-bit 数）
}

// registerToFC 把 thing.Poller.Register 字符串映射成 FC
func registerToFC(reg string) (byte, bool) {
	switch reg {
	case "coil":
		return FCReadCoils, true
	case "discrete_input":
		return FCReadDiscreteInputs, true
	case "holding_register":
		return FCReadHoldingRegisters, true
	case "input_register":
		return FCReadInputRegisters, true
	}
	return 0, false
}

// newModbusPoller 从显式 thing.Poller 创建一个 ModbusPoller
// FC 优先级：
//  1. thing.Poller.Register 非空 → 直接映射（以 Register 为准）
//  2. 否则从 Address 推断（显式 "FC3:xxx" 或数字前缀）
func newModbusPoller(proto string, p thing.Poller) (*ModbusPoller, error) {
	spec, err := ParseAddr(p.Address, TypeUint16, p.Length)
	if err != nil {
		return nil, fmt.Errorf("modbus(%s) Poller Address=%q: %w", proto, p.Address, err)
	}

	// Register 字段优先 —— 如果显式指定了 Register，覆盖 Address 推断的 FC
	if fc, ok := registerToFC(p.Register); ok {
		spec.FC = fc
	}

	return &ModbusPoller{
		fc:        spec.FC,
		startAddr: spec.StartAddr,
		quantity:  spec.Quantity,
	}, nil
}

// calcPollerQuantity 计算区间的读取数量
// bit 类型：last.StartAddr - begin.StartAddr + 1
// 寄存器类型：last.StartAddr + registerQuantity(last.DataType) - begin.StartAddr
func calcPollerQuantity(begin, last *MapItem, isBitType bool) uint16 {
	if isBitType {
		return last.StartAddr - begin.StartAddr + 1
	}
	lastWords := uint16(registerQuantity(last.DataType))
	return last.StartAddr + lastWords - begin.StartAddr
}

// autoGeneratePollers 从 MapItem 列表自动生成 Pollers
//
// 规则：
//  1. 按 FC 分组（coils=FC1, discrete=FC2, holding=FC3, input=FC4）
//  2. 每组按 StartAddr 升序排序
//  3. 相邻点位地址差 ≤ gap 阈值（bit 类型 = 1+sep；寄存器类型 = word 数+sep）视为连续，合并成一个区间
//  4. 区间长度：bit 类型 = last.StartAddr - begin.StartAddr + 1；寄存器类型 = last.StartAddr + word - begin.StartAddr
func autoGeneratePollers(items []MapItem) []ModbusPoller {
	if len(items) == 0 {
		return nil
	}

	// 1. 按 FC 分桶
	buckets := make(map[byte][]*MapItem)
	for i := range items {
		it := &items[i]
		buckets[it.FC] = append(buckets[it.FC], it)
	}

	var pollers []ModbusPoller

	// 2. 对每个 FC bucket 排序 + 合并
	for fc, fcItems := range buckets {
		sort.Slice(fcItems, func(i, j int) bool {
			return fcItems[i].StartAddr < fcItems[j].StartAddr
		})

		isBitType := fc == FCReadCoils || fc == FCReadDiscreteInputs

		var begin, last *MapItem
		for _, it := range fcItems {
			if begin == nil {
				begin = it
				last = it
				continue
			}

			// 比较 next item 的起始地址 vs 当前区间允许的最大间隔
			var cutOff uint16
			if isBitType {
				cutOff = last.StartAddr + 1 + pollerSep
			} else {
				words := uint16(registerQuantity(last.DataType))
				cutOff = last.StartAddr + words + pollerSep
			}

			if it.StartAddr > cutOff {
				// 切断 — 输出 [begin, last] 区间
				pollers = append(pollers, ModbusPoller{
					fc:        fc,
					startAddr: begin.StartAddr,
					quantity:  calcPollerQuantity(begin, last, isBitType),
				})
				begin = it
			}
			last = it
		}

		if begin != nil {
			pollers = append(pollers, ModbusPoller{
				fc:        fc,
				startAddr: begin.StartAddr,
				quantity:  calcPollerQuantity(begin, last, isBitType),
			})
		}
	}

	return pollers
}

// ============================================================
// ModbusPoller 方法
// ============================================================

// FC 返回功能码
func (p *ModbusPoller) FC() byte { return p.fc }

// StartAddr 返回起始地址（0-based）
func (p *ModbusPoller) StartAddr() uint16 { return p.startAddr }

// Quantity 返回读取数量
func (p *ModbusPoller) Quantity() uint16 { return p.quantity }
