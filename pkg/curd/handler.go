package curd

import (
	"reflect"
	"strconv"
	"time"

	"github.com/god-jason/iot-master/pkg/db"

	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

type Handler[T any] struct {
	model T

	// 表名（xorm Table）
	Collection string

	// 可更新字段
	UpdateFields []string

	// 搜索关键词对应的字段名（用于 $or 多字段 LIKE）
	Keywords []string

	// 默认过滤条件（gin context key → 过滤字段名）
	Filter map[string]string

	// IsInt64ID 标记主键为 int64 自增
	IsInt64ID bool
}

// defaultFilterCond 从 gin context 提取字段值，返回 []builder.Cond
func (h *Handler[T]) defaultFilterCond(c *gin.Context) []builder.Cond {
	var conds []builder.Cond
	for fieldName, ctxKey := range h.Filter {
		if val, ok := c.Get(ctxKey); ok {
			conds = append(conds, builder.Eq{fieldName: val})
		}
	}
	return conds
}

func (h *Handler[T]) Search(c *gin.Context) {
	var params ListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		Error(c, 400, "无效的请求参数")
		return
	}
	params.Normalize()

	conds := []builder.Cond{}

	// 列表参数里的过滤条件
	if c := params.BuildCond(); c != nil {
		conds = append(conds, c)
	}

	// 关键词多字段 OR
	if params.HasKeyword() && len(h.Keywords) > 0 {
		kws := make([]builder.Cond, 0, len(h.Keywords))
		for _, kw := range h.Keywords {
			kws = append(kws, builder.Like{kw, "%" + params.Keyword + "%"})
		}
		conds = append(conds, builder.Or(kws...))
	}

	// context 注入的默认过滤
	conds = append(conds, h.defaultFilterCond(c)...)

	var cond builder.Cond
	if len(conds) > 1 {
		cond = builder.And(conds...)
	} else if len(conds) == 1 {
		cond = conds[0]
	}

	total, err := db.Count(h.Collection, cond)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}

	orderAsc := params.Order != "desc"
	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "id"
	}

	result := make([]T, 0, params.PageSize)
	_, err = db.PagedFind(h.Collection, cond, params.Page, params.PageSize, orderBy, orderAsc, &result)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}

	SuccessTotal(c, result, total)
}

func (h *Handler[T]) Get(c *gin.Context) {
	idVal, ok := h.parseID(c.Param("id"))
	if !ok {
		Error(c, 400, "无效的ID")
		return
	}

	conds := []builder.Cond{builder.Eq{"id": idVal}}
	for _, c2 := range h.defaultFilterCond(c) {
		conds = append(conds, c2)
	}
	var cond builder.Cond = builder.And(conds...)

	result := new(T)
	found, err := db.FindOne(h.Collection, cond, result)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}
	if !found || isZero(result) {
		Error(c, 404, "数据不存在")
		return
	}

	Success(c, result)
}

func (h *Handler[T]) Create(c *gin.Context) {
	result := new(T)
	if err := c.ShouldBindJSON(result); err != nil {
		Error(c, 400, "无效的请求参数")
		return
	}

	setID(result)
	setCreated(result)

	if len(h.Filter) > 0 {
		h.setDefaultFilterValues(result, c)
	}

	_, err := db.InsertOne(h.Collection, result)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}

	Success(c, result)
}

func (h *Handler[T]) Update(c *gin.Context) {
	idVal, ok := h.parseID(c.Param("id"))
	if !ok {
		Error(c, 400, "无效的ID")
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		Error(c, 400, "无效的请求参数")
		return
	}

	updateSet := map[string]interface{}{}
	for k, v := range updateData {
		if k == "id" || k == "created" || k == "created_at" {
			continue
		}
		if str, ok := v.(string); ok && str == "" {
			// 空字符串字段跳过
			continue
		}
		if len(h.UpdateFields) > 0 {
			allowed := false
			for _, field := range h.UpdateFields {
				if k == field {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		updateSet[k] = v
	}

	if len(updateSet) == 0 {
		Error(c, 400, "没有可更新的字段")
		return
	}

	conds := []builder.Cond{builder.Eq{"id": idVal}}
	for _, c2 := range h.defaultFilterCond(c) {
		conds = append(conds, c2)
	}
	cond := builder.And(conds...)

	affected, err := db.UpdateOne(h.Collection, cond, updateSet)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		Error(c, 404, "数据不存在")
		return
	}

	Success(c, nil)
}

func (h *Handler[T]) Delete(c *gin.Context) {
	idVal, ok := h.parseID(c.Param("id"))
	if !ok {
		Error(c, 400, "无效的ID")
		return
	}

	conds := []builder.Cond{builder.Eq{"id": idVal}}
	for _, c2 := range h.defaultFilterCond(c) {
		conds = append(conds, c2)
	}
	cond := builder.And(conds...)

	affected, err := db.DeleteOne(h.Collection, cond)
	if err != nil {
		Error(c, 500, err.Error())
		return
	}
	if affected == 0 {
		Error(c, 404, "数据不存在")
		return
	}

	Success(c, nil)
}

// ============ 辅助 ============

func isZero(obj interface{}) bool {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.IsZero()
}

func setID(obj interface{}) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.Kind() == reflect.String && idField.String() == "" {
		idField.SetString(db.NewID())
	}
}

func setCreated(obj interface{}) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	createdField := val.FieldByName("Created")
	if createdField.IsValid() && createdField.Type() == reflect.TypeOf(time.Time{}) {
		createdField.Set(reflect.ValueOf(time.Now()))
	}
}

func (h *Handler[T]) setDefaultFilterValues(obj interface{}, c *gin.Context) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()
	for fieldName, ctxKey := range h.Filter {
		if ctxVal, ok := c.Get(ctxKey); ok {
			field := findFieldByJsonTag(typ, fieldName)
			if field != nil {
				fieldVal := val.FieldByIndex(field.Index)
				if fieldVal.IsValid() && fieldVal.CanSet() {
					setFieldValue(fieldVal, ctxVal)
				}
			}
		}
	}
}

func findFieldByJsonTag(typ reflect.Type, tagName string) *reflect.StructField {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if commaIdx := indexOf(jsonTag, ','); commaIdx > 0 {
				if jsonTag[:commaIdx] == tagName {
					return &field
				}
			}
		}
		if field.Type.Kind() == reflect.Struct && !field.Anonymous {
			if nestedField := findFieldByJsonTag(field.Type, tagName); nestedField != nil {
				return &reflect.StructField{
					Name:   field.Name + "." + nestedField.Name,
					Type:   nestedField.Type,
					Index:  append(field.Index, nestedField.Index...),
					Tag:    nestedField.Tag,
					Offset: field.Offset + nestedField.Offset,
				}
			}
		}
	}
	return nil
}

func setFieldValue(fieldVal reflect.Value, val interface{}) {
	switch fieldVal.Type().Kind() {
	case reflect.String:
		if str, ok := val.(string); ok {
			fieldVal.SetString(str)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := val.(type) {
		case int:
			fieldVal.SetInt(int64(v))
		case int64:
			fieldVal.SetInt(v)
		case string:
			if num, err := strconv.ParseInt(v, 10, 64); err == nil {
				fieldVal.SetInt(num)
			}
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			fieldVal.SetBool(b)
		}
	}
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// parseID 将路径参数 string 转为正确类型
func (h *Handler[T]) parseID(raw string) (interface{}, bool) {
	if h.IsInt64ID {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, false
		}
		return n, true
	}
	return raw, true
}
