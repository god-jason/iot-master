package table

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/rs/xid"
	"xorm.io/builder"
)

// Insert 插入数据
func (t *Table) Insert(values map[string]any) (id any, err error) {
	if len(values) == 0 {
		err = errors.New("没有要插入的数据")
		return
	}

	//删除冗余字段
	for k := range values {
		column := t.Column(k)
		if column == nil {
			delete(values, k)
		}
	}

	var increment bool

	for _, column := range t.Fields {
		//查询自增主键
		if column.Primary && column.Increment {
			increment = true
		}

		//主键，生成默认ID
		if column.Primary && column.Name == "id" && column.Type == "string" {
			if val, ok := values[column.Name]; ok {
				id = val //记录原ID

				if v, ok := val.(string); ok && v == "" {
					id = xid.New().String()
					values[column.Name] = id
				}
			} else {
				id = xid.New().String()
				values[column.Name] = id
			}
		}

		if column.Created {
			values[column.Name] = time.Now().Format(time.DateTime) //直接格式化
		}

		if column.Updated {
			values[column.Name] = time.Now().Format(time.DateTime) //直接格式化
		}

		if column.Json {
			values[column.Name], _ = json.Marshal(values[column.Name])
		}

		if column.Type == "datetime" {
			if val, ok := values[column.Name]; ok {
				values[column.Name] = normalizeDateTimeValue(val)
			}
		}
	}

	if t.BeforeInsert != nil {
		err = t.BeforeInsert(values)
		if err != nil {
			return
		}
	}

	var vs []interface{}
	for k, v := range values {
		vs = append(vs, builder.Eq{db.Engine().Quote(k): v})
	}
	bdr := builder.Dialect(db.Engine().DriverName()).Insert(vs...).Into(db.Engine().Quote(t.Name))
	res, err := db.Engine().Exec(bdr)
	if err != nil {
		return id, err
	}

	//获取自增ID
	if increment {
		id, err = res.LastInsertId()
	}

	if t.AfterInsert != nil {
		err = t.AfterInsert(id, values)
	}

	return
}