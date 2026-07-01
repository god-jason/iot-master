package table

import (
	"github.com/god-jason/iot-master/pkg/api"
)

// Join 关联查询配置
type Join struct {
	Table   string            `json:"table,omitempty"`   //表名
	Alias   string            `json:"alias,omitempty"`   //表别名，默认t1,t2,t3...
	Local   string            `json:"local,omitempty"`   //主表字段
	Foreign string            `json:"foreign,omitempty"` //附表字段（外键）
	Fields  map[string]string `json:"fields,omitempty"`  //字段映射 field=>as
}

// ParamSearch 搜索参数
type ParamSearch struct {
	Skip   int            `form:"skip" json:"skip"`     //越过条数
	Limit  int            `form:"limit" json:"limit"`   //限制条数
	Sort   map[string]int `form:"sort" json:"sort"`     //排序 仅支持一个字段
	Filter map[string]any `form:"filter" json:"filter"` //条件
	Joins  []*Join        `form:"joins" json:"joins"`   //联合查询的字段
	Fields []string       `form:"fields" json:"fields"` //要查询的字段
}

// Aggregator 聚合函数定义
type Aggregator struct {
	Func  string `json:"func"`  // 函数名: count, sum, avg, max, min
	Field string `json:"field"` // 字段名
	As    string `json:"as"`    // 结果别名
}

// ParamGroup 聚合查询参数
type ParamGroup struct {
	By          []string       `json:"by"`          // 分组字段列表
	Aggregators []*Aggregator  `json:"aggregators"` // 聚合函数列表
	Filter      map[string]any `json:"filter"`      // 过滤条件
	Joins       []*Join        `json:"joins"`       // 关联查询配置
	Sort        map[string]int `json:"sort"`        // 排序
	Limit       int            `json:"limit"`       // 限制条数
	Having      map[string]any `json:"having"`      // HAVING 条件（可选）
}

func init() {
	//表接口 普通权限
	api.Register("POST", "table/:table/count", ApiCount)
	api.Register("POST", "table/:table/group", ApiGroup)
	api.Register("POST", "table/:table/create", ApiCreate)
	api.Register("PUT", "table/:table/create", ApiCreate)
	api.Register("POST", "table/:table/update/*id", ApiUpdate)
	api.Register("GET", "table/:table/delete/*id", ApiDelete)
	api.Register("DELETE", "table/:table/delete/*id", ApiDelete)
	api.Register("GET", "table/:table/detail/*id", ApiDetail)
	api.Register("GET", "table/:table/query/*id", ApiQuery)
	api.Register("POST", "table/:table/search", ApiSearch)
	api.Register("POST", "table/:table/import", ApiImport)
	api.Register("POST", "table/:table/export", ApiExport)
}
