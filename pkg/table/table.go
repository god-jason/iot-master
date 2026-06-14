package table

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

// normalizeDateTimeValue 标准化日期时间值
// 支持多种日期格式的自动转换
func normalizeDateTimeValue(value any) any {
	str, ok := value.(string)
	if !ok {
		return value
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return value
	}

	// Already in MySQL DATETIME format.
	if t, err := time.ParseInLocation(time.DateTime, str, time.Local); err == nil {
		return t.Format(time.DateTime)
	}

	// Support date-only values from date picker.
	if t, err := time.ParseInLocation("2006-01-02", str, time.Local); err == nil {
		return t.Format(time.DateTime)
	}

	// Support RFC3339/RFC3339Nano values from frontend ISO strings.
	if t, err := time.Parse(time.RFC3339Nano, str); err == nil {
		return t.Local().Format(time.DateTime)
	}

	return value
}

// ColumnToCondition 将字段和值转换为查询条件
// 支持的操作符: >, >=, <, <=, =, !=, ~, %
func ColumnToCondition(f *Field, val string, hasJoin bool) (cond builder.Cond, err error) {
	fn := db.Engine().Quote(f.Name)
	if hasJoin {
		fn = "t." + db.Engine().Quote(f.Name)
	}

	var v any
	switch val[0] {
	case '>':
		if val[1] == '=' {
			v, err = f.Cast(val[2:])
			if err != nil {
				return
			}
			cond = builder.Gte{fn: v}
		} else {
			v, err = f.Cast(val[1:])
			if err != nil {
				return
			}
			cond = builder.Gt{fn: v}
		}
	case '<':
		if val[1] == '=' {
			v, err = f.Cast(val[2:])
			if err != nil {
				return
			}
			cond = builder.Lte{fn: v}
		} else {
			v, err = f.Cast(val[1:])
			if err != nil {
				return
			}
			cond = builder.Lt{fn: v}
		}
	case '=': //此处冗余了
		if val[1] == '=' {
			v, err = f.Cast(val[2:])
			if err != nil {
				return
			}
			cond = builder.Eq{fn: v}
		} else {
			v, err = f.Cast(val[1:])
			if err != nil {
				return
			}
			cond = builder.Eq{fn: v}
		}
	case '!', '~':
		if val[1] == '=' {
			v, err = f.Cast(val[2:])
			if err != nil {
				return
			}
			cond = builder.Neq{fn: v}
		} else {
			v, err = f.Cast(val[1:])
			if err != nil {
				return
			}
			cond = builder.Neq{fn: v}
		}
	case '%':
		cond = builder.Like{fn, val} //使用原始
	default:
		//前缀
		if val[len(val)-1] == '%' {
			cond = builder.Like{fn, val} //使用原始
		} else {
			cond = builder.Eq{fn: val}
		}
	}
	return
}

// Table 表结构定义
type Table struct {
	Name          string   `json:"name,omitempty"`           // 表名
	Comment       string   `json:"comment,omitempty"`        // 表注释
	Fields        []*Field `json:"fields,omitempty"`         // 字段列表
	Joins         []*Join  `json:"joins,omitempty"`          // 关联配置
	DisableInsert bool     `json:"disable_insert,omitempty"` // 禁用插入
	DisableUpdate bool     `json:"disable_update,omitempty"` // 禁用更新
	DisableDelete bool     `json:"disable_delete,omitempty"` // 禁用删除

	// 钩子函数
	Hook

	indexedFields map[string]*Field // 字段索引缓存
}

// Init 初始化表结构，编译钩子函数
func (t *Table) Init() error {
	t.indexedFields = make(map[string]*Field)
	for _, column := range t.Fields {
		t.indexedFields[column.Name] = column
	}

	return t.Hook.Compile()
}

// Field 根据名称获取字段
func (t *Table) Column(name string) *Field {
	return t.indexedFields[name]
}

// PrimaryKeys 获取所有主键字段
func (t *Table) PrimaryKeys() []*Field {
	var columns []*Field
	for _, column := range t.Fields {
		if column.Primary {
			columns = append(columns, column)
		}
	}
	return columns
}

// condId 根据ID生成查询条件
func (t *Table) condId(id any) (conds []builder.Cond, err error) {
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
		conds = append(conds, builder.Eq{column.Name: val})
	} else {
		//多主键的情况
		if str, ok := id.(string); !ok {
			return nil, errors.New("多主键id需是string类型")
		} else {
			ss := strings.Split(str, "/")
			if len(ss) != len(keys) {
				return nil, errors.New("主键数量不匹配")
			}

			for i, column := range keys {
				val, err := column.Cast(ss[i])
				if err != nil {
					return nil, err
				}
				conds = append(conds, builder.Eq{db.Engine().Quote(column.Name): val})
			}
		}
	}

	return
}

// condWhere 将过滤条件转换为查询条件
// 支持 $or 和 $and 组合条件
func (t *Table) condWhere(filter map[string]any, hasJoin bool) (conds []builder.Cond, err error) {
	for k, v := range filter {

		//多级查询支持
		if k == "$or" {
			if sub, ok := v.(map[string]any); ok && len(sub) > 0 {
				cs, err := t.condWhere(sub, hasJoin)
				if err != nil {
					return nil, err
				}
				or := builder.Or(cs...)
				conds = append(conds, or)
			}
			continue
		} else if k == "$and" {
			if sub, ok := v.(map[string]any); ok && len(sub) > 0 {
				cs, err := t.condWhere(sub, hasJoin)
				if err != nil {
					return nil, err
				}
				and := builder.And(cs...)
				conds = append(conds, and)
			}
			continue
		}

		column := t.Column(k)
		if column == nil {
			return nil, fmt.Errorf("字段 %s 不存在", k)
		}

		switch val := v.(type) {
		case []string:
			for _, s := range val {
				cond, err := ColumnToCondition(column, s, hasJoin)
				if err != nil {
					return nil, err
				}
				conds = append(conds, cond)
			}
		case string:
			cond, err := ColumnToCondition(column, val, hasJoin)
			if err != nil {
				return nil, err
			}
			conds = append(conds, cond)
		case map[string]any:
			// 支持比较查询 {created:{$gt:"",$lt:""}}
			fn := db.Engine().Quote(column.Name)
			if hasJoin {
				fn = "t." + db.Engine().Quote(column.Name)
			}
			for op, opVal := range val {
				switch op {
				case "$gt", ">":
					if str, ok := opVal.(string); ok && str != "" {
						v, err := column.Cast(str)
						if err != nil {
							return nil, err
						}
						conds = append(conds, builder.Gt{fn: v})
					}
				case "$gte", ">=":
					if str, ok := opVal.(string); ok && str != "" {
						v, err := column.Cast(str)
						if err != nil {
							return nil, err
						}
						conds = append(conds, builder.Gte{fn: v})
					}
				case "$lt", "<":
					if str, ok := opVal.(string); ok && str != "" {
						v, err := column.Cast(str)
						if err != nil {
							return nil, err
						}
						conds = append(conds, builder.Lt{fn: v})
					}
				case "$lte", "<=":
					if str, ok := opVal.(string); ok && str != "" {
						v, err := column.Cast(str)
						if err != nil {
							return nil, err
						}
						conds = append(conds, builder.Lte{fn: v})
					}
				case "$ne", "!=", "~=", "<>":
					if str, ok := opVal.(string); ok && str != "" {
						v, err := column.Cast(str)
						if err != nil {
							return nil, err
						}
						conds = append(conds, builder.Neq{fn: v})
					}
				}
			}
		default:
			fn := db.Engine().Quote(column.Name)
			if hasJoin {
				fn = "t." + db.Engine().Quote(column.Name)
			}
			conds = append(conds, builder.Eq{fn: val})
		}
	}
	return
}

// Schema 构建 xorm 表结构
func (t *Table) Schema() *schemas.Table {
	table := schemas.NewTable(t.Name, nil)
	table.Comment = t.Comment

	// 转化列
	for _, column := range t.Fields {
		col := column.ToColumn()
		table.AddColumn(col)
	}

	return table
}

// AddColumn 添加字段到表中
func (t *Table) AddColumn(column *Field) {
	t.indexedFields[column.Name] = column
	t.Fields = append(t.Fields, column)
}

// Create 创建表
func (t *Table) Create() error {
	schema := t.Schema()

	//创建表
	sql, _, err := db.Engine().Dialect().CreateTableSQL(context.Background(), db.Engine().DB(), schema, t.Name)
	if err != nil {
		return err
	}
	_, err = db.Engine().Exec(sql)
	if err != nil {
		return err
	}

	//创建索引
	for _, index := range schema.Indexes {
		sql := db.Engine().Dialect().CreateIndexSQL(t.Name, index)
		_, err := db.Engine().Exec(sql)
		if err != nil {
			return err
		}
	}

	return nil
}

// Drop 删除表
func (t *Table) Drop() error {
	sql, _ := db.Engine().Dialect().DropTableSQL(t.Name)
	_, err := db.Engine().Exec(sql)
	return err
}
