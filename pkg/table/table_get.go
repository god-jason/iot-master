package table

import (
	"encoding/json"
	"errors"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Get 根据ID获取单条数据
func (t *Table) Get(id any, columns []string) (Document, error) {
	bdr := builder.Dialect(db.Engine().DriverName()).Select(columns...).From(db.Engine().Quote(t.Name))

	cs, err := t.condId(id)
	if err != nil {
		return nil, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	rows, err := db.Engine().QueryInterface(bdr)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("记录不存在")
	}
	row := rows[0]

	//解析
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

	return row, nil
}