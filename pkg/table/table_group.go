package table

import (
	"strings"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/spf13/cast"
	"xorm.io/builder"
)

// Group 聚合查询（基础查询，不带关联）
func (t *Table) Group(body *ParamGroup) (rows []map[string]any, err error) {
	bdr := builder.Dialect(db.Engine().DriverName())

	// 构建查询列
	var columns []string

	// 添加分组字段
	for _, f := range strings.Split(body.GroupBy, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		columns = append(columns, db.Engine().Quote(f))
	}

	// 添加聚合函数
	for _, agg := range body.Aggregators {
		expr := agg.Func + "(" + db.Engine().Quote(agg.Field) + ")"
		if agg.As != "" {
			expr += " AS " + db.Engine().Quote(agg.As)
		}
		columns = append(columns, expr)
	}

	bdr.Select(columns...).From(db.Engine().Quote(t.Name))

	// 添加过滤条件
	cs, err := t.condWhere(body.Filter, false)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	// 添加 GROUP BY（直接使用字符串）
	if body.GroupBy != "" {
		bdr.GroupBy(body.GroupBy)
	}

	// 添加 HAVING 条件（可选）
	if len(body.Having) > 0 {
		for k, v := range body.Having {
			switch val := v.(type) {
			case string:
				if len(val) > 0 {
					if val[0] == '>' {
						if len(val) > 1 && val[1] == '=' {
							bdr.Having(db.Engine().Quote(k) + " >= " + val[2:])
						} else {
							bdr.Having(db.Engine().Quote(k) + " > " + val[1:])
						}
					} else if val[0] == '<' {
						if len(val) > 1 && val[1] == '=' {
							bdr.Having(db.Engine().Quote(k) + " <= " + val[2:])
						} else {
							bdr.Having(db.Engine().Quote(k) + " < " + val[1:])
						}
					} else if val[0] == '=' {
						bdr.Having(db.Engine().Quote(k) + " = " + val[1:])
					} else if val[0] == '!' && len(val) > 1 && val[1] == '=' {
						bdr.Having(db.Engine().Quote(k) + " != " + val[2:])
					} else {
						bdr.Having(db.Engine().Quote(k) + " = " + val)
					}
				}
			case int, int64, float64:
				bdr.Having(db.Engine().Quote(k) + " = " + cast.ToString(val))
			case map[string]any:
				for op, opVal := range val {
					switch op {
					case "$gt", ">":
						bdr.Having(db.Engine().Quote(k) + " > " + cast.ToString(opVal))
					case "$gte", ">=":
						bdr.Having(db.Engine().Quote(k) + " >= " + cast.ToString(opVal))
					case "$lt", "<":
						bdr.Having(db.Engine().Quote(k) + " < " + cast.ToString(opVal))
					case "$lte", "<=":
						bdr.Having(db.Engine().Quote(k) + " <= " + cast.ToString(opVal))
					case "$ne", "!=", "~=":
						bdr.Having(db.Engine().Quote(k) + " != " + cast.ToString(opVal))
					case "$eq", "=":
						bdr.Having(db.Engine().Quote(k) + " = " + cast.ToString(opVal))
					}
				}
			}
		}
	}

	// 添加排序
	if len(body.Sort) > 0 {
		for k, v := range body.Sort {
			f := db.Engine().Quote(k)
			if v > 0 {
				bdr.OrderBy(f + " ASC")
			} else {
				bdr.OrderBy(f + " DESC")
			}
		}
	}

	// 添加限制
	if body.Limit > 0 {
		bdr.Limit(body.Limit)
	}

	rows, err = db.Engine().QueryInterface(bdr)
	return
}