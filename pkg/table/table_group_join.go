package table

import (
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/spf13/cast"
	"xorm.io/builder"
)

// GroupJoin 聚合查询（支持关联查询）
func (t *Table) GroupJoin(body *ParamGroup) (rows []map[string]any, err error) {
	joins := body.Joins
	if len(joins) == 0 {
		joins = t.Joins // 默认使用表定义的关联
	}

	// 无关联时使用基础 Group
	if len(joins) == 0 && body.GroupBy != "" {
		return t.Group(body)
	}

	bdr := builder.Dialect(db.Engine().DriverName())

	// 构建查询列
	var columns []string

	// 添加分组字段（带表别名）
	for _, f := range strings.Split(body.GroupBy, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		// 判断是否是关联表字段（格式：table.field 或直接 field）
		if strings.Contains(f, ".") {
			columns = append(columns, f)
		} else {
			columns = append(columns, "t."+db.Engine().Quote(f))
		}
	}

	// 添加聚合函数
	for _, agg := range body.Aggregators {
		field := agg.Field
		// 判断字段来源表
		if strings.Contains(field, ".") {
			field = field // 已指定表别名
		} else {
			field = "t." + db.Engine().Quote(field)
		}
		expr := agg.Func + "(" + field + ")"
		if agg.As != "" {
			expr += " AS " + db.Engine().Quote(agg.As)
		}
		columns = append(columns, expr)
	}

	// 添加 JOIN 表的字段
	for i, join := range joins {
		as := "t" + strconv.Itoa(i+1)
		if join.Field != "" {
			fieldExpr := as + "." + db.Engine().Quote(join.Field)
			if join.As != "" {
				fieldExpr += " AS " + db.Engine().Quote(join.As)
			}
			columns = append(columns, fieldExpr)
		}
	}

	// 主表
	bdr.Select(columns...).From(builder.As(db.Engine().Quote(t.Name), "t"))

	// 添加过滤条件（WHERE）
	cs, err := t.condWhere(body.Filter, true)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	// 添加 JOIN
	for i, join := range joins {
		as := "t" + strconv.Itoa(i+1)
		lf := "t." + db.Engine().Quote(join.LocalField)
		ff := as + "." + db.Engine().Quote(join.ForeignField)
		bdr.LeftJoin(builder.As(db.Engine().Quote(join.Table), as), lf+"="+ff)
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