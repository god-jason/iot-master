package table

// 表处理模块
// 提供数据库表的增删改查、关联查询等操作

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/rs/xid"
	"github.com/spf13/cast"
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
	// 构建xorm schema
	// var table schemas.Table
	//table.Name = t.Name
	//table.Comment = t.Comment

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
	// 第二个参数 checkIfExist 没有处理
	sql, _ := db.Engine().Dialect().DropTableSQL(t.Name)
	_, err := db.Engine().Exec(sql)
	return err
}

// Insert 插入数据
func (t *Table) Insert(values map[string]any) (id any, err error) {
	if len(values) == 0 {
		err = errors.New("没有要插入的数据")
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

	//_, err = db.Engine().Table(t.Name).Insert(values) 原始方式

	if t.AfterInsert != nil {
		err = t.AfterInsert(id, values)
	}

	return
}

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

	//bdr.Where(builder.Eq{"id": id})
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

// Delete 根据条件删除数据
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

// DeleteById 根据ID删除数据
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

// DeleteByIdEx 根据ID删除数据（带额外过滤条件）
func (t *Table) DeleteByIdEx(id any, filter map[string]any) (rows int64, err error) {
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

	if len(filter) > 0 {
		cs, err = t.condWhere(filter, false)
		if err != nil {
			return 0, err
		}
		for _, c := range cs {
			bdr.Where(c)
		}
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

// Join 查询数据列表（支持关联查询）
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
	for i, join := range joins {
		//body.Fields = append(body.Fields)
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

	for i, join := range joins {
		as := "t" + strconv.Itoa(i+1)
		lf := "t." + db.Engine().Quote(join.LocalField)
		ff := as + "." + db.Engine().Quote(join.ForeignField)
		//bdr.LeftJoin(builder.As(join.Table, as), builder.Eq{lf: ff})
		bdr.LeftJoin(builder.As(db.Engine().Quote(join.Table), as), lf+"="+ff) //Eq会转化成字符串参数
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

// Get 根据ID获取单条数据
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

// Count 统计符合条件的数据数量
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
		return 0, errors.New("没有可计数的值")
	}

	for _, v := range res[0] {
		return cast.ToInt64(v), nil
	}

	return 0, errors.New("没有可计数的值")
}

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
		lf := "t." + db.Engine().Quote(join.LocalField)
		columns = append(columns, lf)
		as := "t" + strconv.Itoa(i+1)
		ff := as + "." + db.Engine().Quote(join.Field)
		columns = append(columns, ff+" AS "+db.Engine().Quote(join.As))
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
		as := "t" + strconv.Itoa(i+1)
		lf := "t." + db.Engine().Quote(join.LocalField)
		ff := as + "." + db.Engine().Quote(join.ForeignField)
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
