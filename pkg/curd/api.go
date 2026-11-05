package curd

import (
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

// Response 统一响应结构
type Response struct {
	Code  int         `json:"code"`
	Msg   string      `json:"message"`
	Data  interface{} `json:"data,omitempty"`
	Total int64       `json:"total,omitempty"`
}

// FilterCondition 过滤条件
type FilterCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符: eq, ne, gt, lt, gte, lte, like, in
	Value    interface{} `json:"value"`    // 值
}

// ListParams 列表查询参数
type ListParams struct {
	Page     int               `form:"page" json:"page"`           // 页码，默认1
	PageSize int               `form:"page_size" json:"page_size"` // 每页数量，默认20，最大100
	Keyword  string            `form:"keyword" json:"keyword"`     // 关键字搜索
	Filter   []FilterCondition `form:"filter" json:"filter"`       // 过滤条件
	OrderBy  string            `form:"order_by" json:"order_by"`   // 排序字段，默认id
	Order    string            `form:"order" json:"order"`         // 排序方向：asc/desc，默认desc
}

// Normalize 标准化列表参数
func (l *ListParams) Normalize() {
	if l.Page <= 0 {
		l.Page = 1
	}
	if l.PageSize <= 0 {
		l.PageSize = 20
	}
	if l.PageSize > 100 {
		l.PageSize = 100
	}
	if l.OrderBy == "" {
		l.OrderBy = "id"
	}
	if l.Order == "" {
		l.Order = "desc"
	}
	if l.Order != "asc" && l.Order != "desc" {
		l.Order = "desc"
	}
}

// Offset 计算偏移量
func (l *ListParams) Offset() int {
	return (l.Page - 1) * l.PageSize
}

// HasKeyword 是否有关键字搜索
func (l *ListParams) HasKeyword() bool {
	return l.Keyword != ""
}

// HasFilter 是否有过滤条件
func (l *ListParams) HasFilter() bool {
	return len(l.Filter) > 0
}

// BuildCond 直接构建 xorm builder.Cond（替换旧的 BuildFilter + ProcessFilterValues）
// 返回 nil 表示无过滤条件
func (l *ListParams) BuildCond() builder.Cond {
	var conds []builder.Cond

	for _, f := range l.Filter {
		cond := condFromOp(f.Operator, f.Field, f.Value)
		if cond != nil {
			conds = append(conds, cond)
		}
	}

	if len(conds) == 0 {
		return nil
	}
	if len(conds) == 1 {
		return conds[0]
	}
	return builder.And(conds...)
}

// condFromOp 将 FilterCondition.Operator 映射为 builder.Cond
func condFromOp(op, field string, val interface{}) builder.Cond {
	switch op {
	case "eq", "":
		return builder.Eq{field: val}
	case "ne":
		return builder.Neq{field: val}
	case "gt":
		return builder.Gt{field: val}
	case "lt":
		return builder.Lt{field: val}
	case "gte":
		return builder.Gte{field: val}
	case "lte":
		return builder.Lte{field: val}
	case "like":
		s, _ := val.(string)
		return builder.Like{field, "%" + s + "%"}
	case "in":
		arr, ok := val.([]interface{})
		if !ok {
			return nil
		}
		return builder.In(field, arr...)
	}
	return nil
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

// SuccessTotal 成功响应（带总数）
func SuccessTotal(c *gin.Context, data interface{}, total int64) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data, Total: total})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}

// ErrorWithStatus 带HTTP状态码的错误响应
func ErrorWithStatus(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, Response{Code: code, Msg: msg})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, msg string) {
	Error(c, 400, msg)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, msg string) {
	ErrorWithStatus(c, http.StatusUnauthorized, 401, msg)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, msg string) {
	ErrorWithStatus(c, http.StatusForbidden, 403, msg)
}

// NotFound 404错误
func NotFound(c *gin.Context, msg string) {
	ErrorWithStatus(c, http.StatusNotFound, 404, msg)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, msg string) {
	Error(c, 500, msg)
}

// ParamError 参数错误
func ParamError(c *gin.Context, msg string) {
	Error(c, 400, msg)
}

// DataNotFound 数据不存在
func DataNotFound(c *gin.Context, msg string) {
	ErrorWithStatus(c, http.StatusNotFound, 404, msg)
}

// ParseJSON 解析JSON请求体
func ParseJSON(c *gin.Context, obj interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(body, obj)
}

// ParseQuery 解析查询参数到结构体
func ParseQuery(c *gin.Context, obj interface{}) error {
	values := c.Request.URL.Query()
	return MapToStruct(values, obj)
}

// MapToStruct 将map转换为struct
func MapToStruct(m map[string][]string, obj interface{}) error {
	data := make(map[string]interface{})
	for k, v := range m {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	jsonBytes, err := sonic.Marshal(data)
	if err != nil {
		return err
	}
	return sonic.Unmarshal(jsonBytes, obj)
}

// ToJSON 将对象转为JSON字符串
func ToJSON(v interface{}) string {
	bytes, _ := sonic.MarshalString(v)
	return bytes
}

// FromJSON 将JSON字符串转为对象
func FromJSON(jsonStr string, v interface{}) error {
	return sonic.UnmarshalString(jsonStr, v)
}
