// Package memory tsdb 回退驱动：进程内内存存储。
// tsdb 未启用、或驱动名配置为 "memory" 时，tsdb.New 会用本实现。
// 适合单元测试与单机极简场景，进程重启数据丢失。
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

type store struct {
	mu    sync.RWMutex
	data  map[string][]tsdbapi.Point
}

func New(cfg map[string]any) (tsdbapi.Store, error) {
	return &store{data: map[string][]tsdbapi.Point{}}, nil
}

func makeKey(measurement string, tags map[string]string) string {
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := measurement
	for _, k := range keys {
		out += "|" + k + "=" + tags[k]
	}
	return out
}

func (m *store) Write(ctx context.Context, pt tsdbapi.Point) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := makeKey(pt.Measurement, pt.Tags)
	m.data[k] = append(m.data[k], pt)
	return nil
}

func (m *store) WriteBatch(ctx context.Context, pts []tsdbapi.Point) error {
	for _, p := range pts {
		if err := m.Write(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (m *store) Query(ctx context.Context, spec tsdbapi.QuerySpec) (map[string]tsdbapi.Series, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []tsdbapi.Point
	for _, pts := range m.data {
		for _, pt := range pts {
			if pt.Measurement != spec.Measurement {
				continue
			}
			if !matchTags(pt.Tags, spec.Tags) {
				continue
			}
			if !spec.Start.IsZero() && pt.Time.Before(spec.Start) {
				continue
			}
			if !spec.End.IsZero() && pt.Time.After(spec.End) {
				continue
			}
			matched = append(matched, pt)
		}
	}
	if len(matched) == 0 {
		return map[string]tsdbapi.Series{}, nil
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].Time.Before(matched[j].Time) })

	buckets := groupByWindow(matched, spec.Window)
	method := spec.Method
	if method == "" {
		method = "mean"
	}

	out := map[string]tsdbapi.Series{}
	for _, b := range buckets {
		fieldSet := map[string]struct{}{}
		for _, pt := range b {
			for k := range pt.Fields {
				fieldSet[k] = struct{}{}
			}
		}
		for f := range fieldSet {
			if len(spec.Fields) > 0 {
				found := false
				for _, want := range spec.Fields {
					if want == f {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			vals := collectNumeric(b, f)
			agg := aggregate(vals, method)
			out[f] = tsdbapi.Series{
				Field:  f,
				Values: append(out[f].Values, agg),
				Times:  append(out[f].Times, b[0].Time.Unix()),
			}
		}
	}
	return out, nil
}

func (m *store) Last(ctx context.Context, measurement string, tags map[string]string) (map[string]tsdbapi.LastResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	k := makeKey(measurement, tags)
	pts, ok := m.data[k]
	if !ok || len(pts) == 0 {
		return nil, nil
	}
	var last tsdbapi.Point
	for _, p := range pts {
		if p.Time.After(last.Time) {
			last = p
		}
	}
	out := map[string]tsdbapi.LastResult{}
	for f, v := range last.Fields {
		out[f] = tsdbapi.LastResult{Time: last.Time.Unix(), Value: v}
	}
	return out, nil
}

func (m *store) RawQuery(ctx context.Context, q string, args ...any) ([]map[string]any, error) {
	return nil, fmt.Errorf("memory store 不支持 RawQuery")
}

func (m *store) Close() error { return nil }

func matchTags(got, want map[string]string) bool {
	for k, v := range want {
		if g, ok := got[k]; !ok || g != v {
			return false
		}
	}
	return true
}

func groupByWindow(pts []tsdbapi.Point, window string) [][]tsdbapi.Point {
	if window == "" || window == "raw" {
		return [][]tsdbapi.Point{pts}
	}
	d, err := time.ParseDuration(window)
	if err != nil || d <= 0 {
		return [][]tsdbapi.Point{pts}
	}
	var out [][]tsdbapi.Point
	var cur []tsdbapi.Point
	var curStart time.Time
	for _, p := range pts {
		if cur == nil || p.Time.After(curStart.Add(d)) {
			if cur != nil {
				out = append(out, cur)
			}
			cur = []tsdbapi.Point{p}
			curStart = p.Time.Truncate(d)
		} else {
			cur = append(cur, p)
		}
	}
	if cur != nil {
		out = append(out, cur)
	}
	return out
}

func collectNumeric(pts []tsdbapi.Point, field string) []float64 {
	var out []float64
	for _, pt := range pts {
		v, ok := pt.Fields[field]
		if !ok {
			continue
		}
		switch n := v.(type) {
		case float64:
			out = append(out, n)
		case float32:
			out = append(out, float64(n))
		case int:
			out = append(out, float64(n))
		case int64:
			out = append(out, float64(n))
		}
	}
	return out
}

func aggregate(vals []float64, method string) any {
	if len(vals) == 0 {
		return nil
	}
	switch method {
	case "count":
		return float64(len(vals))
	case "sum":
		var s float64
		for _, v := range vals {
			s += v
		}
		return s
	case "max":
		m := vals[0]
		for _, v := range vals[1:] {
			if v > m {
				m = v
			}
		}
		return m
	case "min":
		m := vals[0]
		for _, v := range vals[1:] {
			if v < m {
				m = v
			}
		}
		return m
	default:
		var s float64
		for _, v := range vals {
			s += v
		}
		return s / float64(len(vals))
	}
}
