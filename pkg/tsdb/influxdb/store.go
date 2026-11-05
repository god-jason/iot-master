// Package influxdb tsdb 驱动：InfluxDB v2（纯 Go，无需 CGO）
//
// New() 返回实现了 tsdbapi.Store 的实例，由 tsdb.New 分派调用。
package influxdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

type store struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	queryAPI api.QueryAPI
	org      string
	bucket   string
}

// New 构造 influxdb Store。cfg 键：url、token、org、bucket
func New(cfg map[string]any) (tsdbapi.Store, error) {
	url := stringVal(cfg, "url")
	if url == "" {
		return nil, fmt.Errorf("influxdb: url 为空")
	}
	token := stringVal(cfg, "token")
	org := stringVal(cfg, "org")
	bucket := stringVal(cfg, "bucket")

	client := influxdb2.NewClient(url, token)
	return &store{
		client:   client,
		writeAPI: client.WriteAPI(org, bucket),
		queryAPI: client.QueryAPI(org),
		org:      org,
		bucket:   bucket,
	}, nil
}

func stringVal(cfg map[string]any, k string) string {
	if v, ok := cfg[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (s *store) Write(ctx context.Context, pt tsdbapi.Point) error {
	tags := make(map[string]string, len(pt.Tags))
	for k, v := range pt.Tags {
		tags[k] = v
	}
	p := influxdb2.NewPoint(pt.Measurement, tags, pt.Fields, pt.Time)
	s.writeAPI.WritePoint(p)
	return nil
}

func (s *store) WriteBatch(ctx context.Context, pts []tsdbapi.Point) error {
	for _, pt := range pts {
		if err := s.Write(ctx, pt); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Query(ctx context.Context, spec tsdbapi.QuerySpec) (map[string]tsdbapi.Series, error) {
	q := s.buildFlux(spec)
	result, err := s.queryAPI.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	out := map[string]tsdbapi.Series{}
	for result.Next() {
		rec := result.Record()
		f := rec.Field()
		out[f] = tsdbapi.Series{
			Field:  f,
			Values: append(out[f].Values, rec.Value()),
			Times:  append(out[f].Times, rec.Time().Unix()),
		}
	}
	return out, result.Err()
}

func (s *store) Last(ctx context.Context, measurement string, tags map[string]string) (map[string]tsdbapi.LastResult, error) {
	var b strings.Builder
	fmt.Fprintf(&b, `from(bucket: "%s")`, s.bucket)
	fmt.Fprintf(&b, "\n  |> range(start: -30d)")
	fmt.Fprintf(&b, "\n  |> filter(fn: (r) => r[\"_measurement\"] == %q)", measurement)
	for k, v := range tags {
		fmt.Fprintf(&b, "\n  |> filter(fn: (r) => r[\"%s\"] == %q)", k, v)
	}
	b.WriteString("\n  |> last()")

	result, err := s.queryAPI.Query(ctx, b.String())
	if err != nil {
		return nil, err
	}
	out := map[string]tsdbapi.LastResult{}
	for result.Next() {
		rec := result.Record()
		out[rec.Field()] = tsdbapi.LastResult{Time: rec.Time().Unix(), Value: rec.Value()}
	}
	return out, result.Err()
}

func (s *store) RawQuery(ctx context.Context, q string, args ...any) ([]map[string]any, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("influxdb Flux 不支持占位符参数")
	}
	result, err := s.queryAPI.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	for result.Next() {
		rec := result.Record()
		row := map[string]any{
			"_time":       rec.Time(),
			"_field":      rec.Field(),
			"_value":      rec.Value(),
			"_measurement": rec.Measurement(),
		}
		for k, v := range rec.Values() {
			row[k] = v
		}
		rows = append(rows, row)
	}
	return rows, result.Err()
}

func (s *store) Close() error {
	if s.writeAPI != nil {
		s.writeAPI.Flush()
	}
	if s.client != nil {
		s.client.Close()
	}
	return nil
}

func (s *store) buildFlux(spec tsdbapi.QuerySpec) string {
	var b strings.Builder
	fmt.Fprintf(&b, `from(bucket: "%s")`, s.bucket)

	start := spec.Start
	end := spec.End
	if start.IsZero() {
		start = time.Now().Add(-24 * time.Hour)
	}
	if end.IsZero() {
		end = time.Now()
	}
	fmt.Fprintf(&b, "\n  |> range(start: %s, stop: %s)",
		start.Format(time.RFC3339), end.Format(time.RFC3339))

	fmt.Fprintf(&b, "\n  |> filter(fn: (r) => r[\"_measurement\"] == %q)", spec.Measurement)
	for k, v := range spec.Tags {
		fmt.Fprintf(&b, "\n  |> filter(fn: (r) => r[\"%s\"] == %q)", k, v)
	}
	if len(spec.Fields) > 0 {
		fmt.Fprintf(&b, "\n  |> filter(fn: (r) => r[\"_field\"] == %q", spec.Fields[0])
		for _, f := range spec.Fields[1:] {
			fmt.Fprintf(&b, " or r[\"_field\"] == %q", f)
		}
		b.WriteString(")")
	}

	window := spec.Window
	if window == "" {
		dur := end.Sub(start)
		switch {
		case dur > 70*24*time.Hour:
			window = "1h"
		case dur > 7*24*time.Hour:
			window = "10m"
		case dur > 14*time.Hour:
			window = "1m"
		case dur > 5000*time.Second:
			window = "10s"
		default:
			window = "1s"
		}
	}
	method := spec.Method
	if method == "" {
		method = "mean"
	}
	fmt.Fprintf(&b, "\n  |> aggregateWindow(every: %s, fn: %s, createEmpty: false)", window, method)
	b.WriteString("\n  |> yield()")
	return b.String()
}
