package table

import (
	"github.com/gin-gonic/gin"
)

type Join struct {
	Table        string `json:"table,omitempty"`         //表名
	LocalField   string `json:"local_field,omitempty"`   //主表字段
	ForeignField string `json:"foreign_field,omitempty"` //附表字段（外键）
	Field        string `json:"field,omitempty"`         //取字段 TODO 可以改为数组
	As           string `json:"as,omitempty"`            //赋值
}

type ParamSearch struct {
	Skip   int            `form:"skip" json:"skip"`     //越过条数
	Limit  int            `form:"limit" json:"limit"`   //限制条数
	Sort   map[string]int `form:"sort" json:"sort"`     //排序 仅支持一个字段
	Filter map[string]any `form:"filter" json:"filter"` //条件
	Joins  []*Join        `form:"joins" json:"joins"`   //联合查询的字段
	Fields []string       `form:"fields" json:"fields"` //要查询的字段
}

func ApiSearch(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}
	var body ParamSearch
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		Error(ctx, err)
		return
	}

	//多租户过滤
	tid := ctx.GetString("tenant")
	if tid != "" {
		column := table.Column("tenant_id")
		if column != nil {
			//只有未传值tenant_id时，才会赋值用户所在的tenant_id
			if _, ok := body.Filter["tenant_id"]; !ok {
				body.Filter["tenant_id"] = tid
			}
		}
	}

	cnt, err := table.Count(body.Filter)
	if err != nil {
		Error(ctx, err)
		return
	}

	results, err := table.Join(&body)
	if err != nil {
		Error(ctx, err)
		return
	}

	//OK(ctx, results)
	List(ctx, results, cnt)
}
