package db

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"xorm.io/builder"
)

// ============ 基础 CRUD ============

// InsertOne 插入单条记录，返回自增主键。
// xorm 1.3.x 的 Insert 返回值是受影响行数，而非 last insert id；
// 但传入结构体指针时 xorm 会把自增 ID 回填到结构体，因此这里从结构体读取真实 ID。
func InsertOne(table string, document interface{}) (int64, error) {
	if _, err := Engine.Table(table).Insert(document); err != nil {
		return 0, err
	}
	rv := reflect.Indirect(reflect.ValueOf(document))
	if rv.IsValid() && rv.Kind() == reflect.Struct {
		if f := rv.FieldByName("ID"); f.IsValid() && f.CanInt() && f.Int() > 0 {
			return f.Int(), nil
		}
	}
	return 0, nil
}

// FindOne 查询单条记录，cond 为空时不加 WHERE
func FindOne(table string, cond builder.Cond, result interface{}) (bool, error) {
	sess := Engine.Table(table)
	if cond != nil {
		sess = sess.Where(cond)
	}
	return sess.Get(result)
}

// Find 查询多条记录（page=0 表示不分页）
func Find(table string, cond builder.Cond, result interface{}, page, pageSize int) error {
	sess := Engine.Table(table)
	if cond != nil {
		sess = sess.Where(cond)
	}
	if page > 0 && pageSize > 0 {
		sess = sess.Limit(pageSize, (page-1)*pageSize)
	}
	return sess.Find(result)
}

// UpdateOne 更新单条记录；update map 的 key 若为 "id"/"created"/"created_at" 会被跳过
func UpdateOne(table string, cond builder.Cond, update map[string]interface{}) (int64, error) {
	if cond == nil {
		return 0, fmt.Errorf("更新条件不能为空")
	}
	cleaned := map[string]interface{}{}
	for k, v := range update {
		if k == "id" || k == "created" || k == "created_at" {
			continue
		}
		cleaned[k] = normalizeUpdateValue(v)
	}
	if len(cleaned) == 0 {
		return 0, nil
	}
	return Engine.Table(table).Where(cond).Cols(mapKeys(cleaned)...).Update(cleaned)
}

// normalizeUpdateValue 将 map 形式的更新值转换为数据库可写入的类型。
// xorm 只对结构体字段按 `json` tag 做序列化，map 形式的更新会原样传参，
// 因此这里统一把 map/slice/struct 编码成 JSON 字符串，避免 "unsupported type" 报错。
func normalizeUpdateValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch v.(type) {
	case []byte, time.Time:
		return v
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
		data, err := json.Marshal(v)
		if err != nil {
			return v
		}
		return string(data)
	default:
		return v
	}
}

// DeleteOne 删除单条记录
func DeleteOne(table string, cond builder.Cond) (int64, error) {
	if cond == nil {
		return 0, fmt.Errorf("删除条件不能为空")
	}
	return Engine.Table(table).Where(cond).Delete()
}

// DeleteMany 删除多条记录
func DeleteMany(table string, cond builder.Cond) (int64, error) {
	if cond == nil {
		return 0, fmt.Errorf("删除条件不能为空")
	}
	return Engine.Table(table).Where(cond).Delete()
}

// Count 统计记录数
func Count(table string, cond builder.Cond) (int64, error) {
	sess := Engine.Table(table)
	if cond != nil {
		sess = sess.Where(cond)
	}
	return sess.Count()
}

// PagedFind 分页查询
func PagedFind(table string, cond builder.Cond, page, pageSize int, orderBy string, orderAsc bool, result interface{}) (int64, error) {
	total, err := Count(table, cond)
	if err != nil {
		return 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	sess := Engine.Table(table)
	if cond != nil {
		sess = sess.Where(cond)
	}
	if orderBy != "" {
		dir := " DESC"
		if orderAsc {
			dir = " ASC"
		}
		sess = sess.OrderBy(orderBy + dir)
	}
	sess = sess.Limit(pageSize, (page-1)*pageSize)
	if err = sess.Find(result); err != nil {
		return 0, err
	}
	return total, nil
}

// Exists 判断是否存在
func Exists(table string, cond builder.Cond) (bool, error) {
	n, err := Count(table, cond)
	return n > 0, err
}

// ============ Upsert ============

// Upsert 有则更新，无则插入
func Upsert(table string, cond builder.Cond, create map[string]interface{}, update map[string]interface{}) (int64, error) {
	n, _ := Count(table, cond)
	if n > 0 {
		return UpdateOne(table, cond, update)
	}
	doc := map[string]interface{}{}
	if cond != nil {
		doc["id"] = nil
	}
	for k, v := range create {
		if k != "id" {
			doc[k] = v
		}
	}
	if v, ok := create["id"]; ok && v != nil && v != "" {
		doc["id"] = v
	}
	return InsertOne(table, doc)
}

// ============ 原生 SQL ============

func RawMaps(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := Engine.SQL(sqlStr, args...).Query()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		m := make(map[string]interface{}, len(row))
		for k, v := range row {
			m[k] = stringOrValue(v)
		}
		results[i] = m
	}
	return results, nil
}

func ExecSQL(sqlStr string, args ...interface{}) (int64, error) {
	res, err := Engine.SQL(sqlStr, args...).Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func ExecRawCount(sqlStr string, args ...interface{}) (int64, error) {
	rows, err := Engine.SQL(sqlStr, args...).Query()
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 || len(rows[0]) == 0 {
		return 0, nil
	}
	for _, v := range rows[0] {
		if n, ok := stringToInt(v); ok {
			return n, nil
		}
	}
	return 0, nil
}

func stringOrValue(v interface{}) interface{} {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func stringToInt(v interface{}) (int64, bool) {
	switch x := v.(type) {
	case []byte:
		var n int64
		_, err := fmt.Sscanf(string(x), "%d", &n)
		return n, err == nil
	case int64:
		return x, true
	case int:
		return int64(x), true
	}
	return 0, false
}

// ============ ID 生成 ============

// NewID 生成 24 位 string 主键
func NewID() string {
	id := uuid.New().String()
	return id[:8] + id[9:13] + id[14:18] + id[19:23] + id[24:28] + id[29:33]
}

// NewSecret 生成 32 字符随机 hex 密钥
func NewSecret() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
