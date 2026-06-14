package table

import (
	"encoding/json"
	"slices"
	"strconv"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Join 查询数据列表（支持关联查询）
func (t *Table) Join(body *ParamSearch) (rows []map[string]any, err error) {
	joins := body.Joins
	if len(joins) == 0 {
		joins = t.Joins //默认使用表定义的关联
	}
	if len(joins) == 0 {
		return t.Find(body) //直接使用基础
	}

	bdr := builder.Dialect(db.Engine().DriverName())

	var columns []string

	if len(body.Fields) == 0 {
		columns = append(columns, "t.*")
	} else {
		for _, f := range body.Fields {
			columns = append(columns, "t."+db.Engine().Quote(f))
		}
	}
	for i, join := range joins {
		if slices.Index(body.Fields, join.LocalField) < 0 {
			lf := "t." + db.Engine().Quote(join.LocalField)
			columns = append(columns, lf)
		}
		as := "t" + strconv.Itoa(i+1)
		ff := as + "." + db.Engine().Quote(join.Field)
		columns = append(columns, ff+" AS "+db.Engine().Quote(join.As))
	}

	bdr.Select(columns...).From(builder.As(db.Engine().Quote(t.Name), "t"))

	cs, err := t.condWhere(body.Filter, true)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	for i, join := range joins {
		as := "t" + strconv.Itoa(i+1)
		lf := "t." + db.Engine().Quote(join.LocalField)
		ff := as + "." + db.Engine().Quote(join.ForeignField)
		bdr.LeftJoin(builder.As(db.Engine().Quote(join.Table), as), lf+"="+ff)
	}

	//排序
	if len(body.Sort) > 0 {
		for k, v := range body.Sort {
			f := "t." + db.Engine().Quote(k)
			if v > 0 {
				bdr.OrderBy(f + " ASC")
			} else {
				bdr.OrderBy(f + " DESC")
			}
		}
	}

	if body.Limit <= 0 {
		body.Limit = 20
	}
	bdr.Limit(body.Limit, body.Skip)

	rows, err = db.Engine().QueryInterface(bdr)
	if err != nil {
		return
	}

	//解析JSON
	for _, row := range rows {
		for _, column := range t.Fields {
			if column.Json {
				if val, ok := row[column.Name]; ok {
					if str, ok := val.(string); ok {
						var v any
						err = json.Unmarshal([]byte(str), &v)
						if err == nil {
							row[column.Name] = v
						}
					}
				}
			}
		}
	}

	return rows, nil
}