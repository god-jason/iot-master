package table

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
)

// Detail 根据id获取单个文档，支持关联查询
func (t *Table) Detail(id any, joins []*Join) (Document, error) {
	// 如果没有传入关联，使用表定义的关联
	if len(joins) == 0 {
		joins = t.Joins
	}
	// 如果没有关联，直接使用 Get
	if len(joins) == 0 {
		return t.Get(id, nil)
	}

	bdr := builder.Dialect(db.Engine().DriverName())

	var columns []string
	columns = append(columns, "t.*")

	for i, join := range joins {
		lf := "t." + db.Engine().Quote(join.Local)
		columns = append(columns, lf)
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

	// 添加 id 条件（带表别名 t.）
	keys := t.PrimaryKeys()
	if len(keys) == 0 {
		column := t.Column("id")
		if column != nil {
			keys = append(keys, column)
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("表没有主键")
	}
	if len(keys) == 1 {
		column := keys[0]
		val, err := column.Cast(id)
		if err != nil {
			return nil, err
		}
		bdr.Where(builder.Eq{"t." + db.Engine().Quote(column.Name): val})
	} else {
		// 多主键的情况
		str, ok := id.(string)
		if !ok {
			return nil, errors.New("多主键id需是string类型")
		}
		ss := strings.Split(str, "/")
		if len(ss) != len(keys) {
			return nil, errors.New("主键数量不匹配")
		}
		for i, column := range keys {
			val, err := column.Cast(ss[i])
			if err != nil {
				return nil, err
			}
			bdr.Where(builder.Eq{"t." + db.Engine().Quote(column.Name): val})
		}
	}

	// 添加关联
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

	// 只查询一条
	bdr.Limit(1, 0)

	rows, err := db.Engine().QueryInterface(bdr)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("记录不存在")
	}
	row := rows[0]

	// 解析 JSON
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
