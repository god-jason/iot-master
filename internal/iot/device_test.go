package iot

import (
	"testing"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/thing"
)

// 验证 BaseDevice.ReportProperties 的物模型校验：
// 1) Property Min < Max 时越界丢弃
// 2) 任意一条 Filter 不通过丢弃
func TestReportPropertiesModelValidation(t *testing.T) {
	Models().Set("m1", &thing.Model{
		Properties: []thing.Property{
			{Key: "temperature", Min: 0, Max: 100},
			{Key: "humidity", Min: 10, Max: 10}, // Min==Max 不检查
		},
		Filters: []thing.Filter{
			{Name: "positive", Expr: "temperature > 0"},
		},
	})
	defer Models().Delete("m1")

	dev := NewBaseDevice(&types.Device{ID: "d1", ModelID: "m1"})

	// 1. Min/Max 越界 → 整条丢弃（实时状态也不更新）
	if err := dev.ReportProperties(map[string]interface{}{"temperature": 120.0, "humidity": 50.0}, time.Now()); err != nil {
		t.Fatalf("ReportProperties: %v", err)
	}
	if _, ok := dev.Property("temperature"); ok {
		t.Fatalf("越界数据不应更新实时状态")
	}

	// 2. Filter 不通过 → 丢弃
	if err := dev.ReportProperties(map[string]interface{}{"temperature": -5.0}, time.Now()); err != nil {
		t.Fatalf("ReportProperties: %v", err)
	}
	if _, ok := dev.Property("temperature"); ok {
		t.Fatalf("Filter 不通过的数据不应更新实时状态")
	}

	// 3. Min==Max 的属性不做范围检查；Filter 通过 → 正常更新
	if err := dev.ReportProperties(map[string]interface{}{"temperature": 25.0, "humidity": 50.0}, time.Now()); err != nil {
		t.Fatalf("ReportProperties: %v", err)
	}
	if v, ok := dev.Property("temperature"); !ok || v != 25.0 {
		t.Fatalf("合格数据应更新实时状态，got %v ok=%v", v, ok)
	}
	if v, ok := dev.Property("humidity"); !ok || v != 50.0 {
		t.Fatalf("Min==Max 时不应做范围检查，got %v ok=%v", v, ok)
	}
}
