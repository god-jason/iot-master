// Package modbus 内部 —— 唯一的地址/FC 解析权威实现
//
// 支持的地址格式：
//
//	"40001" / "400001"  → FC3 (holding)    StartAddr = N - 40001
//	"30001" / "300001"  → FC4 (input)      StartAddr = N - 30001
//	"00001" / "000001"  → FC1 (coil)       StartAddr = N - 1
//	"10001" / "100001"  → FC2 (discrete)   StartAddr = N - 10001
//	"FC3:100"           → 显式 FC + 显式 StartAddr
//	"FC3:100:10"        → 显式 FC + StartAddr + Quantity
package modbus

import (
	"fmt"
	"strconv"
	"strings"
)

// AddrSpec 解析结果 —— 唯一输出结构
type AddrSpec struct {
	FC        byte
	StartAddr uint16
	Quantity  uint16
}

// ParseAddr 唯一入口 —— 解析寄存器地址
//
// 参数：
//
//	address   地址字符串（必填）
//	dt        DataType enum（推断 Quantity 用）
//	length    显式数量（>0 时覆盖 dt 推断）
//
// FC 推断优先级：
//
//  1. address 中 "FC3:xxx" 显式前缀
//  2. address 数字前缀推断（4xxxx → FC3, 3xxxx → FC4, 0xxxx → FC1, 1xxxx → FC2）
func ParseAddr(address string, dt DataType, length int) (*AddrSpec, error) {
	spec := &AddrSpec{Quantity: 1}

	upAddr := strings.ToUpper(strings.TrimSpace(address))

	// Step 1: 显式 FC 前缀 "FC3:100" / "FC3:100:10"
	if strings.HasPrefix(upAddr, "FC") {
		parts := strings.Split(upAddr, ":")
		fcByte, err := strconv.ParseUint(strings.TrimPrefix(parts[0], "FC"), 10, 8)
		if err != nil {
			return nil, fmt.Errorf("modbus: 无法解析 FC %q: %w", parts[0], err)
		}
		spec.FC = byte(fcByte)

		start, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("modbus: 无法解析 StartAddr %q: %w", parts[1], err)
		}
		spec.StartAddr = uint16(start)

		qty := uint16(0)
		if len(parts) >= 3 {
			if q, err := strconv.ParseUint(parts[2], 10, 16); err == nil && q > 0 {
				qty = uint16(q)
			}
		}
		spec.Quantity = resolveQuantity(qty, dt, length, spec.FC)
		return spec, nil
	}

	// Step 2: 数字前缀推断
	addrNum, err := strconv.ParseUint(strings.TrimSpace(address), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("modbus: 地址 %q 既不是显式 FC 也不是数字: %w", address, err)
	}
	switch {
	case addrNum >= 40001 && addrNum <= 49999:
		spec.FC = FCReadHoldingRegisters
		spec.StartAddr = uint16(addrNum - 40001)
	case addrNum >= 30001 && addrNum <= 39999:
		spec.FC = FCReadInputRegisters
		spec.StartAddr = uint16(addrNum - 30001)
	case addrNum >= 1 && addrNum <= 9999:
		spec.FC = FCReadCoils
		spec.StartAddr = uint16(addrNum - 1)
	case addrNum >= 10001 && addrNum <= 19999:
		spec.FC = FCReadDiscreteInputs
		spec.StartAddr = uint16(addrNum - 10001)
	default:
		return nil, fmt.Errorf("modbus: 地址 %q 前缀无法推断 FC (期望 4xxxx / 3xxxx / 0xxxx / 1xxxx)", address)
	}

	spec.Quantity = resolveQuantity(0, dt, length, spec.FC)
	return spec, nil
}

// resolveQuantity 统一 Quantity 推断逻辑
func resolveQuantity(qty uint16, dt DataType, length int, fc byte) uint16 {
	if qty > 0 {
		return qty
	}
	if length > 0 {
		return uint16(length)
	}
	if fc == FCReadCoils || fc == FCReadDiscreteInputs {
		return 1
	}
	return registerQuantity(dt)
}
