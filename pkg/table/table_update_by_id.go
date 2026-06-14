package table

import (
	"encoding/json"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// UpdateById 根据ID更新数据
func (t *Table) UpdateById(id any, values map[string]any) (rows int64, err error) {
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

	if t.BeforeUpdate != nil {
		err = t.BeforeUpdate(id, values)
		if err != nil {
			return
		}
	}

	var updates []builder.Cond
	for k, v := range values {
		updates = append(updates, builder.Eq{db.Engine().Quote(k): v})
	}
	bdr := builder.Dialect(db.Engine().DriverName()).Update(updates...).From(db.Engine().Quote(t.Name))

	cs, err := t.condId(id)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	res, err := db.Engine().ID(id).Exec(bdr)
	if err != nil {
		return 0, err
	}

	if t.AfterUpdate != nil {
		err = t.AfterUpdate(id, values, values)
	}

	return res.RowsAffected()
}

// UpdateByIdEx 根据ID更新数据（带额外过滤条件）
func (t *Table) UpdateByIdEx(id any, filter map[string]any, values map[string]any) (rows int64, err error) {
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

	if t.BeforeUpdate != nil {
		err = t.BeforeUpdate(id, values)
		if err != nil {
			return
		}
	}

	var updates []builder.Cond
	for k, v := range values {
		updates = append(updates, builder.Eq{db.Engine().Quote(k): v})
	}
	bdr := builder.Dialect(db.Engine().DriverName()).Update(updates...).From(db.Engine().Quote(t.Name))

	cs, err := t.condId(id)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	if len(filter) > 0 {
		cs, err = t.condWhere(filter, false)
		if err != nil {
			return 0, err
		}
		for _, c := range cs {
			bdr.Where(c)
		}
	}

	res, err := db.Engine().ID(id).Exec(bdr)
	if err != nil {
		return 0, err
	}

	if t.AfterUpdate != nil {
		err = t.AfterUpdate(id, values, values)
	}

	return res.RowsAffected()
}