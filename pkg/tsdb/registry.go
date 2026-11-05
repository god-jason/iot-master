package tsdb

import (
	"fmt"
	"sort"

	"github.com/god-jason/iot-master/pkg/tsdb/influxdb"
	"github.com/god-jason/iot-master/pkg/tsdb/memory"
	"github.com/god-jason/iot-master/pkg/tsdb/tdengine"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

// Drivers 返回可用驱动名（排序后）
func Drivers() []string { return []string{"influxdb", "memory", "tdengine"} }

// New 按驱动名创建 Store。
// cfg 由 boot.Startup 按 driver 类型填充（见 boot.go）。
func New(driver string, cfg map[string]any) (tsdbapi.Store, error) {
	switch driver {
	case "influxdb":
		return influxdb.New(cfg)
	case "tdengine":
		return tdengine.New(cfg)
	case "", "memory":
		return memory.New(cfg)
	default:
		names := Drivers()
		sort.Strings(names)
		return nil, fmt.Errorf("tsdb: 未注册的驱动 %q（可用: %v）", driver, names)
	}
}
