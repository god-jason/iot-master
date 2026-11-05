package ota

import (
	"encoding/json"
	"fmt"
)

// Strategy 灰度策略
//
//	{"type":"canary","success_threshold":0.95,"failure_threshold":0.10,
//	 "batches":[{"percent":5,"wait_minutes":30},{"percent":20,"wait_minutes":60},{"percent":100,"wait_minutes":0}]}
type Strategy struct {
	Type             string          `json:"type"`              // canary / all
	SuccessThreshold float64         `json:"success_threshold"` // 批次成功率门禁
	FailureThreshold float64         `json:"failure_threshold"` // 批次失败率上限
	Batches          []StrategyBatch `json:"batches"`           // 批次定义
}

// StrategyBatch 单个批次定义
type StrategyBatch struct {
	Percent     int `json:"percent"`      // 目标百分比
	WaitMinutes int `json:"wait_minutes"` // 上一批完成后的等待分钟数
}

// BatchPlan 批次切分结果
type BatchPlan struct {
	Seq         int // 批次序号，从 1 开始
	Percent     int // 策略中的百分比
	WaitMinutes int // 等待分钟数
	Count       int // 实际设备数
}

// DefaultStrategy 默认策略：一次性全量下发
func DefaultStrategy() Strategy {
	return Strategy{
		Type:             "all",
		SuccessThreshold: DefaultSuccessThreshold(),
		FailureThreshold: DefaultFailureThreshold(),
		Batches:          []StrategyBatch{{Percent: 100, WaitMinutes: 0}},
	}
}

// NormalizeStrategy 补全阈值、过滤非法批次并保证至少一个批次
func NormalizeStrategy(s *Strategy) {
	if s == nil {
		return
	}
	if s.Type == "" {
		s.Type = "canary"
	}
	if s.SuccessThreshold <= 0 || s.SuccessThreshold > 1 {
		s.SuccessThreshold = DefaultSuccessThreshold()
	}
	if s.FailureThreshold <= 0 || s.FailureThreshold > 1 {
		s.FailureThreshold = DefaultFailureThreshold()
	}

	cleaned := make([]StrategyBatch, 0, len(s.Batches))
	for _, b := range s.Batches {
		if b.Percent <= 0 {
			continue
		}
		if b.Percent > 100 {
			b.Percent = 100
		}
		if b.WaitMinutes < 0 {
			b.WaitMinutes = 0
		}
		cleaned = append(cleaned, b)
	}
	if len(cleaned) == 0 {
		cleaned = []StrategyBatch{{Percent: 100, WaitMinutes: 0}}
	}
	s.Batches = cleaned
}

// ParseStrategy 解析数据库中的策略 JSON，失败时返回默认策略
func ParseStrategy(raw string) Strategy {
	if raw == "" {
		return DefaultStrategy()
	}
	var s Strategy
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return DefaultStrategy()
	}
	NormalizeStrategy(&s)
	return s
}

// BuildBatchPlans 按百分比切分设备；前 n-1 批按比例取整，余量全部归最后一批。
func BuildBatchPlans(total int, s Strategy) []BatchPlan {
	if total <= 0 {
		return nil
	}
	NormalizeStrategy(&s)
	plans := make([]BatchPlan, 0, len(s.Batches))
	assigned := 0
	for i, b := range s.Batches {
		plan := BatchPlan{Seq: i + 1, Percent: b.Percent, WaitMinutes: b.WaitMinutes}
		if i == len(s.Batches)-1 {
			plan.Count = total - assigned
		} else {
			plan.Count = total * b.Percent / 100
			if plan.Count < 1 {
				plan.Count = 1
			}
			if plan.Count > total-assigned {
				plan.Count = total - assigned
			}
		}
		assigned += plan.Count
		plans = append(plans, plan)
		if assigned >= total {
			break
		}
	}
	// 极端情况下（前面批次已覆盖全部）保证最后一个计划有效
	if len(plans) == 0 {
		plans = []BatchPlan{{Seq: 1, Percent: 100, Count: total}}
	}
	return plans
}

// ValidateStrategy 校验策略是否可执行
func ValidateStrategy(s Strategy) error {
	NormalizeStrategy(&s)
	if len(s.Batches) == 0 {
		return fmt.Errorf("至少需要一个批次")
	}
	for i, b := range s.Batches {
		if b.Percent <= 0 {
			return fmt.Errorf("第 %d 批百分比无效", i+1)
		}
	}
	return nil
}
