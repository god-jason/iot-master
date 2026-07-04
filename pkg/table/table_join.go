package table

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

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

	//以下优化无用，真正卡的是计算总数 select count(*) where device_id=? 遍历了所有数据
	//判断是否有单主键，如果有，可以使用JoinFast优化
	//pks := len(t.PrimaryKeys())
	//if pks == 0 {
	//	return t.JoinFast("id", joins, body)
	//}
	//if pks == 1 {
	//	return t.JoinFast(t.PrimaryKeys()[0].Name, joins, body)
	//}

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

	cs, err := t.condWhere(body.Filter, true)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		bdr.Where(c)
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
