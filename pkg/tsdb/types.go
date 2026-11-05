// Package tsdb 时序数据库统一入口。
//
// 驱动选择由 config.driver 决定，registry.New() 用 switch/case 分派子驱动。
// 这种模式（父包 switch 调子驱动，子驱动只 import 独立接口包 tsdbapi）
// 避免了 Go 里"父包 import 子包、子包又 import 父包"的循环依赖。
//
// 子驱动实现见：
//   - pkg/tsdb/influxdb
//   - pkg/tsdb/tdengine
// 默认回退：memory（进程内存，适合测试与单机极简场景）
package tsdb

import (
	"fmt"

	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

// DefaultStore 进程全局 Store，由 boot.Startup 初始化
var DefaultStore tsdbapi.Store

// DSN 把分散字段拼成连接串（供子驱动复用）
// influxdb: dsn 无效，用 url/token/org/bucket
// tdengine: dsn 格式 taosRestful://user:pass@host:port/database
func DSN(driver, username, password, host string, port int, database string) string {
	switch driver {
	case "tdengine":
		return fmt.Sprintf("taosRestful://%s:%s@%s:%d/%s", username, password, host, port, database)
	default:
		return ""
	}
}
