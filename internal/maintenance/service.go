package maintenance

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/god-jason/iot-master/internal/ota"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"xorm.io/builder"
)

// NextDue 计算下一次到期时间：base 按周期类型累加 cycleValue
func NextDue(base time.Time, cycleType string, cycleValue int) time.Time {
	if cycleValue <= 0 {
		cycleValue = 1
	}
	switch cycleType {
	case types.MaintenanceCycleMonth:
		return base.AddDate(0, cycleValue, 0)
	case types.MaintenanceCycleYear:
		return base.AddDate(cycleValue, 0, 0)
	default: // 按天（未识别的周期类型同样按天处理）
		return base.AddDate(0, 0, cycleValue)
	}
}

// OrderNo 生成维保工单号：WB + 秒级时间戳 + 3 位随机数
func OrderNo() string {
	return fmt.Sprintf("WB%s%03d", time.Now().Format("20060102150405"), rand.Intn(1000))
}

// ResolveDevices 按目标范围解析设备（复用 OTA 的范围解析逻辑，常量取值一致）
func ResolveDevices(scopeType, scopeID string) ([]types.Device, error) {
	return ota.ResolveDevices(scopeType, scopeID)
}

// GenerateForPlan 为指定计划生成工单：每台设备在当前到期周期内只生成一张工单。
// 返回本次生成数量与目标设备总数；生成完成后顺延计划的下次到期时间。
func GenerateForPlan(plan *types.MaintenancePlan, now time.Time) (int, int, error) {
	if plan == nil {
		return 0, 0, fmt.Errorf("计划不存在")
	}
	if plan.NextDue.IsZero() {
		return 0, 0, fmt.Errorf("计划未设置下次到期时间")
	}

	devices, err := ResolveDevices(plan.ScopeType, plan.ScopeID)
	if err != nil {
		return 0, 0, err
	}

	generated := 0
	for i := range devices {
		device := &devices[i]

		// 同一计划、同一设备、同一到期周期已生成过工单则跳过
		n, err := db.Count("maintenance_order", builder.And(
			builder.Eq{"plan_id": plan.ID},
			builder.Eq{"device_id": device.ID},
			builder.Eq{"due_at": plan.NextDue},
		))
		if err != nil {
			return generated, len(devices), err
		}
		if n > 0 {
			continue
		}

		order := types.MaintenanceOrder{
			OrderNo:    OrderNo(),
			PlanID:     plan.ID,
			DeviceID:   device.ID,
			Type:       types.MaintenanceOrderTypePlan,
			Title:      plan.Name,
			Priority:   2,
			Status:     types.MaintenanceOrderStatusPending,
			AssigneeID: plan.OwnerID,
			DueAt:      plan.NextDue,
			Items:      plan.Items,
			CreatedBy:  plan.CreatedBy,
			Created:    now,
			Updated:    now,
		}
		if _, err := db.InsertOne("maintenance_order", &order); err != nil {
			return generated, len(devices), err
		}
		generated++
	}

	// 顺延下次到期时间
	plan.NextDue = NextDue(plan.NextDue, plan.CycleType, plan.CycleValue)
	plan.LastGenerated = now
	plan.Updated = now
	if _, err := db.UpdateOne("maintenance_plan", builder.Eq{"id": plan.ID}, map[string]interface{}{
		"next_due":       plan.NextDue,
		"last_generated": now,
		"updated":        now,
	}); err != nil {
		return generated, len(devices), err
	}

	return generated, len(devices), nil
}

// GenerateDuePlans 扫描到期（含提前提醒天数）的计划并批量生成工单。
// 单个计划失败只记录日志，不阻断其他计划。
func GenerateDuePlans(now time.Time) (int, error) {
	var plans []types.MaintenancePlan
	if err := db.Find("maintenance_plan", builder.Eq{"disabled": false}, &plans, 0, 0); err != nil {
		return 0, err
	}

	generated := 0
	for i := range plans {
		plan := &plans[i]
		if plan.NextDue.IsZero() {
			continue
		}
		// 到达提醒窗口（next_due - remind_days）后才生成工单
		due := plan.NextDue.AddDate(0, 0, -plan.RemindDays)
		if now.Before(due) {
			continue
		}
		n, total, err := GenerateForPlan(plan, now)
		if err != nil {
			logger.Warn("维保计划生成工单失败", "plan", plan.ID, "error", err)
			continue
		}
		generated += n
		if n > 0 {
			logger.Info("维保计划已生成工单", "plan", plan.ID, "generated", n, "total", total)
		}
	}
	return generated, nil
}
