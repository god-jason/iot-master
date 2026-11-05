// Package tsdbapi 时序数据库统一接口与数据模型（零外部依赖）。
//
// 驱动子包（pkg/tsdb/influxdb、pkg/tsdb/tdengine）实现本包 Store 接口。
// 上层业务只依赖本包接口，不关心底层驱动。
//
// 这是 Go 中避免"父包 import 子包、子包又 import 父包"循环依赖的经典做法，
// 类似标准库 database/sql/driver 与 database/sql 的分层。
package tsdbapi

import (
	"context"
	"time"
)

// Point 一条时序数据点。
// Measurement 相当于 measurement / 超级表名；Tags 是维度；Fields 是数值列。
type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]any
	Time        time.Time
}

// Series 一个 field 的时序序列。
type Series struct {
	Field  string
	Values []any
	Times  []int64 // Unix 秒
}

// QuerySpec 查询规格。
// 不同驱动在 Store.Query 内部各自翻译：influxdb → Flux，tdengine → SQL。
type QuerySpec struct {
	Measurement string
	Tags        map[string]string
	Start       time.Time
	End         time.Time
	Fields      []string // 空或 nil 返回全部 field
	Window      string   // 聚合窗口 "1m"/"10s"；空表示直接取原始
	Method      string   // mean/max/min/sum/count；空默认 mean
	Limit       int
	OrderDesc   bool
}

// LastResult 每个 field 的最新一条
type LastResult struct {
	Time  int64
	Value any
}

// Store 时序存储统一接口
type Store interface {
	Write(ctx context.Context, pt Point) error
	WriteBatch(ctx context.Context, pts []Point) error
	Query(ctx context.Context, spec QuerySpec) (map[string]Series, error)
	Last(ctx context.Context, measurement string, tags map[string]string) (map[string]LastResult, error)
	// RawQuery 原生查询逃生舱，两种驱动语法不同（Flux vs TDengine SQL）
	RawQuery(ctx context.Context, q string, args ...any) ([]map[string]any, error)
	Close() error
}
