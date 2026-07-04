package table

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Join 快速联合查询
func (t *Table) JoinFast(pk string, joins []*Join, body *ParamSearch) (rows []map[string]any, err error) {

	//第一步，分页查询，获得ID集合
	var ids []any
	subBdr := builder.Dialect(db.Engine().DriverName())
	//subBdr.Select("t." + db.Engine().Quote(pk[0].Name)).From(builder.As(db.Engine().Quote(t.Name), "t"))

	cs, err := t.condWhere(body.Filter, false)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		subBdr.Where(c)
	}

	if len(body.Sort) > 0 {
		for k, v := range body.Sort {
			f := db.Engine().Quote(k)
			if v > 0 {
				subBdr.OrderBy(f + " ASC")
			} else {
				subBdr.OrderBy(f + " DESC")
			}
		}
	}

	if body.Limit <= 0 {
		body.Limit = 20
	}
	subBdr.Limit(body.Limit, body.Skip)

	subRows, err := db.Engine().QueryInterface(subBdr)
	if err != nil {
		return nil, err
	}

	for _, row := range subRows {
		if val, ok := row[pk]; ok {
			ids = append(ids, val)
		}
	}

	if len(ids) == 0 {
		return []map[string]any{}, nil
	}

	//第二步，根据ID集合查询关联数据
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
		if slices.Index(body.Fields, join.Local) < 0 {
			lf := "t." + db.Engine().Quote(join.Local)
			columns = append(columns, lf)
		}
		as := join.Alias
		if as == "" {
			as = "t" + strconv.Itoa(i+1)
		}
		for field, alias := range join.Fields {
			if alias == "" {
				alias = join.Table + "_" + field
			}
			ff := as + "." + db.Engine().Quote(field)
			columns = append(columns, ff+" AS "+db.Engine().Quote(alias))
		}
	}

	bdr.Select(columns...).From(builder.As(db.Engine().Quote(t.Name), "t"))

	if len(pk) > 0 {
		bdr.Where(builder.In("t."+db.Engine().Quote(pk), ids))
	}

	for i, join := range joins {
		as := join.Alias
		if as == "" {
			as = "t" + strconv.Itoa(i+1)
		}
		var lf string
		if strings.Contains(join.Local, ".") {
			lf = join.Local
		} else {
			lf = "t." + db.Engine().Quote(join.Local)
		}
		ff := as + "." + db.Engine().Quote(join.Foreign)
		bdr.LeftJoin(builder.As(db.Engine().Quote(join.Table), as), lf+"="+ff)
	}

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
