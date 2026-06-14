package table

import (
	"encoding/json"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Update 根据条件更新数据
func (t *Table) Update(filter map[string]any, values map[string]any) (rows int64, err error) {
	for _, column := range t.Fields {
		if column.Updated {
			values[column.Name] = time.Now().Format(time.DateTime) //直接格式化
		}

		if column.Json {
			if val, ok := values[column.Name]; ok {
				values[column.Name], _ = json.Marshal(val)
			}
		}

		if column.Type == "datetime" {
			if val, ok := values[column.Name]; ok {
				values[column.Name] = normalizeDateTimeValue(val)
			}
		}
	}

	var updates []builder.Cond
	for k, v := range values {
		updates = append(updates, builder.Eq{db.Engine().Quote(k): v})
	}

	bdr := builder.Dialect(db.Engine().DriverName()).Update(updates...).From(db.Engine().Quote(t.Name))

	cs, err := t.condWhere(filter, false)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	res, err := db.Engine().Exec(bdr)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
