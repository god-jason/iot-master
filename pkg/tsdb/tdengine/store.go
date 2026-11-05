// Package tdengine tsdb 驱动：TDengine 3.x REST 模式（纯 Go，无 CGO 依赖）
//
// 使用 taosdata/driver-go/v3 的 taosRestful database/sql 驱动，
// 连接串格式：taosRestful://username:password@host:6041/database
//
// TDengine 强 schema：超级表（STable）在首次写入 measurement 时尝试 CREATE STABLE IF NOT EXISTS。
// 注意 TDengine 不允许后续新增 field 改变 schema——业务应在首次写入前确定完整列集。
package tdengine

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/taosdata/driver-go/v3/taosRestful"

	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

type store struct {
	db       *sql.DB
	database string
	stables  sync.Map
}

// New 构造 TDengine Store。cfg 键：host、port、username、password、database
func New(cfg map[string]any) (tsdbapi.Store, error) {
	host := stringVal(cfg, "host")
	if host == "" {
		return nil, fmt.Errorf("tdengine: host 为空")
	}
	port := intVal(cfg, "port", 6041)
	user := stringVal(cfg, "username")
	if user == "" {
		user = "root"
	}
	pass := stringVal(cfg, "password")
	dbName := stringVal(cfg, "database")
	if dbName == "" {
		dbName = "dmp"
	}

	dsn := fmt.Sprintf("taosRestful://%s:%s@%s:%d/%s", user, pass, host, port, dbName)

	if conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprintf("%d", port)), 2*time.Second); err == nil {
		conn.Close()
	}

	db, err := sql.Open("taosRestful", dsn)
	if err != nil {
		return nil, fmt.Errorf("tdengine sql.Open: %w", err)
	}
	db.SetMaxOpenConns(16)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("tdengine ping: %w", err)
	}
	_, _ = db.ExecContext(context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	_, _ = db.ExecContext(context.Background(), fmt.Sprintf("USE %s", dbName))

	return &store{db: db, database: dbName}, nil
}

func stringVal(cfg map[string]any, k string) string {
	if v, ok := cfg[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
func intVal(cfg map[string]any, k string, def int) int {
	if v, ok := cfg[k]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		}
	}
	return def
}
func quote(n string) string { return "`" + strings.ReplaceAll(n, "`", "``") + "`" }

func (s *store) ensureStable(measurement string, tags []string, fields []string) {
	if _, ok := s.stables.Load(measurement); ok {
		return
	}
	s.stables.Store(measurement, true)

	cols := []string{"ts TIMESTAMP"}
	for _, f := range fields {
		cols = append(cols, fmt.Sprintf("`%s` FLOAT", f))
	}
	tagCols := []string{}
	for _, t := range tags {
		tagCols = append(tagCols, fmt.Sprintf("`%s` VARCHAR(64)", t))
	}

	sqlStmt := fmt.Sprintf("CREATE STABLE IF NOT EXISTS %s (%s)", quote(measurement), strings.Join(cols, ", "))
	if len(tagCols) > 0 {
		sqlStmt += fmt.Sprintf(" TAGS (%s)", strings.Join(tagCols, ", "))
	}
	_, _ = s.db.ExecContext(context.Background(), sqlStmt)
}

func subTableName(measurement string, tags map[string]string) string {
	if len(tags) == 0 {
		return quote(measurement + "_default")
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString(measurement)
	for _, k := range keys {
		sb.WriteString("_")
		sb.WriteString(k)
		sb.WriteString("_")
		sb.WriteString(tags[k])
	}
	return quote(sb.String())
}

func (s *store) Write(ctx context.Context, pt tsdbapi.Point) error {
	tagKeys := make([]string, 0, len(pt.Tags))
	for k := range pt.Tags {
		tagKeys = append(tagKeys, k)
	}
	sort.Strings(tagKeys)
	fieldKeys := make([]string, 0, len(pt.Fields))
	for k := range pt.Fields {
		fieldKeys = append(fieldKeys, k)
	}
	sort.Strings(fieldKeys)

	s.ensureStable(pt.Measurement, tagKeys, fieldKeys)

	cols := []string{"ts"}
	placeholders := []string{"?"}
	args := []any{pt.Time.UnixMilli()}

	for _, f := range fieldKeys {
		cols = append(cols, quote(f))
		placeholders = append(placeholders, "?")
		args = append(args, pt.Fields[f])
	}

	sqlStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		subTableName(pt.Measurement, pt.Tags),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "))
	if len(tagKeys) > 0 {
		tagPh := make([]string, len(tagKeys))
		tagArgs := make([]any, len(tagKeys))
		for i, k := range tagKeys {
			tagPh[i] = "?"
			tagArgs[i] = pt.Tags[k]
		}
		sqlStmt += fmt.Sprintf(" USING %s TAGS (%s)", quote(pt.Measurement), strings.Join(tagPh, ", "))
		args = append(args, tagArgs...)
	}

	_, err := s.db.ExecContext(ctx, sqlStmt, args...)
	return err
}

func (s *store) WriteBatch(ctx context.Context, pts []tsdbapi.Point) error {
	for _, p := range pts {
		if err := s.Write(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Query(ctx context.Context, spec tsdbapi.QuerySpec) (map[string]tsdbapi.Series, error) {
	start, end := spec.Start, spec.End
	if start.IsZero() {
		start = time.Now().Add(-24 * time.Hour)
	}
	if end.IsZero() {
		end = time.Now()
	}

	var where []string
	where = append(where, "ts BETWEEN ? AND ?")
	args := []any{start.UnixMilli(), end.UnixMilli()}
	for k, v := range spec.Tags {
		where = append(where, fmt.Sprintf("`%s` = ?", k))
		args = append(args, v)
	}

	method := spec.Method
	if method == "" {
		method = "AVG"
	}
	window := spec.Window
	if window == "" {
		window = "1s"
	}

	var selectFields []string
	fieldKeys := spec.Fields
	if len(fieldKeys) == 0 {
		selectFields = append(selectFields, "*")
	} else {
		for _, f := range fieldKeys {
			selectFields = append(selectFields, fmt.Sprintf("%s(`%s`) AS `%s`", strings.ToUpper(method), f, f))
		}
	}

	sqlStmt := fmt.Sprintf("SELECT ts, %s FROM %s WHERE %s INTERVAL(%s)",
		strings.Join(selectFields, ", "),
		quote(spec.Measurement),
		strings.Join(where, " AND "),
		window)
	if spec.Limit > 0 {
		sqlStmt += fmt.Sprintf(" LIMIT %d", spec.Limit)
	}
	if spec.OrderDesc {
		sqlStmt += " ORDER BY ts DESC"
	}

	rows, err := s.db.QueryContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	out := map[string]tsdbapi.Series{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		for i, c := range cols {
			if i == 0 {
				continue
			}
			out[c] = tsdbapi.Series{
				Field:  c,
				Values: append(out[c].Values, vals[i]),
				Times:  append(out[c].Times, toUnix(vals[0])),
			}
		}
	}
	return out, rows.Err()
}

func toUnix(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case time.Time:
		return t.Unix()
	default:
		return 0
	}
}

func (s *store) Last(ctx context.Context, measurement string, tags map[string]string) (map[string]tsdbapi.LastResult, error) {
	var sqlStmt string
	if len(tags) > 0 {
		var where []string
		var args []any
		keys := make([]string, 0, len(tags))
		for k := range tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			where = append(where, fmt.Sprintf("`%s` = ?", k))
			args = append(args, tags[k])
		}
		sqlStmt = fmt.Sprintf("SELECT LAST_ROW(*) FROM %s WHERE %s", quote(measurement), strings.Join(where, " AND "))
		rows, err := s.db.QueryContext(ctx, sqlStmt, args...)
		if err != nil {
			return nil, err
		}
		return s.scanLast(rows, tags)
	}
	sqlStmt = fmt.Sprintf("SELECT LAST_ROW(*) FROM %s", quote(measurement))
	rows, err := s.db.QueryContext(ctx, sqlStmt)
	if err != nil {
		return nil, err
	}
	return s.scanLast(rows, tags)
}

func (s *store) scanLast(rows *sql.Rows, tags map[string]string) (map[string]tsdbapi.LastResult, error) {
	defer rows.Close()
	cols, _ := rows.Columns()
	out := map[string]tsdbapi.LastResult{}
	if rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		var ts int64
		for i, c := range cols {
			if i == 0 {
				ts = toUnix(vals[i])
				continue
			}
			if _, ok := tags[c]; ok {
				continue
			}
			out[c] = tsdbapi.LastResult{Time: ts, Value: vals[i]}
		}
	}
	return out, rows.Err()
}

func (s *store) RawQuery(ctx context.Context, q string, args ...any) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, c := range cols {
			row[c] = vals[i]
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (s *store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
