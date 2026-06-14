package table

import (
	"encoding/json"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Find 查询数据列表（基础查询，不带关联）
func (t *Table) Find(body *ParamSearch) (rows []map[string]any, err error) {
	columns := body.Fields
	if len(columns) == 0 {
		columns = []string{"*"}
	}
	bdr := builder.Dialect(db.Engine().DriverName()).Select(columns...).From(db.Engine().Quote(t.Name))

	cs, err := t.condWhere(body.Filter, false)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	//排序
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
	return
}