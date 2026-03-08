package table

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/smart"
	"github.com/rs/xid"
	"github.com/spf13/cast"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

func ColumnToCondition(f *smart.Column, val string, hasJoin bool) (cond builder.Cond, err error) {
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

type Table struct {
	Name          string          `json:"name,omitempty"`
	Comment       string          `json:"comment,omitempty"`
	Columns       []*smart.Column `json:"columns,omitempty"`
	Joins         []*Join         `json:"joins,omitempty"`
	DisableInsert bool            `json:"disable_insert,omitempty"`
	DisableUpdate bool            `json:"disable_update,omitempty"`
	DisableDelete bool            `json:"disable_delete,omitempty"`

	//原生钩子
	Hook

	indexedColumns map[string]*smart.Column
}

func (t *Table) Init() error {
	t.indexedColumns = make(map[string]*smart.Column)
	for _, column := range t.Columns {
		t.indexedColumns[column.Name] = column
	}

	return t.Hook.Compile()
}

func (t *Table) Column(name string) *smart.Column {
	return t.indexedColumns[name]
}

func (t *Table) PrimaryKeys() []*smart.Column {
	var columns []*smart.Column
	for _, column := range t.Columns {
		if column.Primary {
			columns = append(columns, column)
		}
	}
	return columns
}

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
			return nil, fmt.Errorf("column %s not found", k)
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

func (t *Table) Schema() *schemas.Table {
	//构建xorm schema
	//var table schemas.Table
	//table.Name = t.Name
	//table.Comment = t.Comment

	table := schemas.NewTable(t.Name, nil)
	table.Comment = t.Comment

	//转化列
	for _, column := range t.Columns {
		col := column.ToColumn()
		table.AddColumn(col)
	}

	return table
}

func (t *Table) AddColumn(column *smart.Column) {
	t.indexedColumns[column.Name] = column
	t.Columns = append(t.Columns, column)
}

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

func (t *Table) Drop() error {
	//第二个参数 checkIfExist 没有处理
	sql, _ := db.Engine().Dialect().DropTableSQL(t.Name)
	_, err := db.Engine().Exec(sql)
	return err
}

func (t *Table) Insert(values map[string]any) (id any, err error) {
	if len(values) == 0 {
		err = errors.New("no values to insert")
		return
	}

	//删除冗余字段
	for k, _ := range values {
		column := t.Column(k)
		if column == nil {
			delete(values, k)
		}
	}

	var increment bool

	for _, column := range t.Columns {
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

	//_, err = db.Engine().Table(t.Name).Insert(values) 原始方式

	if t.AfterInsert != nil {
		err = t.AfterInsert(id, values)
	}

	return
}

func (t *Table) Update(filter map[string]any, values map[string]any) (rows int64, err error) {
	for _, column := range t.Columns {
		if column.Updated {
			values[column.Name] = time.Now().Format(time.DateTime) //直接格式化
		}

		if column.Json {
			if val, ok := values[column.Name]; ok {
				values[column.Name], _ = json.Marshal(val)
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

func (t *Table) UpdateById(id any, values map[string]any) (rows int64, err error) {
	for _, column := range t.Columns {
		if column.Updated {
			values[column.Name] = time.Now().Format(time.DateTime) //直接格式化
		}

		if column.Json {
			if val, ok := values[column.Name]; ok {
				values[column.Name], _ = json.Marshal(val)
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

	//bdr.Where(builder.Eq{"id": id})
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

func (t *Table) Delete(filter map[string]any) (rows int64, err error) {
	cs, err := t.condWhere(filter, false)
	if err != nil {
		return 0, err
	}

	bdr := builder.Dialect(db.Engine().DriverName()).Delete(cs...).From(db.Engine().Quote(t.Name))

	res, err := db.Engine().Exec(bdr)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (t *Table) DeleteById(id any) (rows int64, err error) {
	bdr := builder.Dialect(db.Engine().DriverName()).Delete().From(db.Engine().Quote(t.Name))

	if t.BeforeDelete != nil {
		err = t.BeforeDelete(id)
		if err != nil {
			return
		}
	}

	cs, err := t.condId(id)
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

	if t.AfterDelete != nil {
		err = t.AfterDelete(id, nil)
		if err != nil {
			return 0, err
		}
	}

	return res.RowsAffected()
}

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
		for _, column := range t.Columns {
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
	for i, join := range body.Joins {
		//body.Columns = append(body.Columns)
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

	for i, join := range body.Joins {
		as := "t" + strconv.Itoa(i+1)
		lf := "t." + db.Engine().Quote(join.LocalField)
		ff := as + "." + db.Engine().Quote(join.ForeignField)
		//bdr.LeftJoin(builder.As(join.Table, as), builder.Eq{lf: ff})
		bdr.LeftJoin(builder.As(join.Table, as), lf+"="+ff) //Eq会转化成字符串参数
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
		for _, column := range t.Columns {
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

func (t *Table) Get(id any, columns []string) (Document, error) {
	bdr := builder.Dialect(db.Engine().DriverName()).Select(columns...).From(db.Engine().Quote(t.Name))

	//bdr.Where(builder.Eq{"id": id})
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
	for _, column := range t.Columns {
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

func (t *Table) Count(filter map[string]any) (cnt int64, err error) {
	bdr := builder.Dialect(db.Engine().DriverName()).Select("count(*)").From(db.Engine().Quote(t.Name))

	cs, err := t.condWhere(filter, false)
	if err != nil {
		return 0, err
	}
	for _, c := range cs {
		bdr.Where(c)
	}

	res, err := db.Engine().QueryInterface(bdr)
	if err != nil {
		return 0, err
	}

	if len(res) == 0 {
		return 0, errors.New("no values to count")
	}

	for _, v := range res[0] {
		return cast.ToInt64(v), nil
	}

	return 0, errors.New("no values to count")
}
