// Package maintenance 维保管理：计划到期扫描、工单自动生成。
package maintenance

import (
	"time"

	"github.com/god-jason/iot-master/pkg/config"
)

// MODULE 维保配置模块名
const MODULE = "maintenance"

func init() {
	config.SetDefault(MODULE, "enabled", true)
	config.SetDefault(MODULE, "scan_interval", "1m")
}

// Enabled 维保模块是否启用
func Enabled() bool { return config.GetBool(MODULE, "enabled") }

// ScanInterval 到期计划扫描间隔
func ScanInterval() time.Duration {
	if s := config.GetString(MODULE, "scan_interval"); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			return d
		}
	}
	if n := config.GetInt(MODULE, "scan_interval"); n > 0 {
		return time.Duration(n) * time.Second
	}
	return time.Minute
}
