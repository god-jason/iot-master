package curd

import (
	"strings"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
)

type Join struct {
	Table        string //表名
	LocaleField  string //主表字段
	ForeignField string //附表字段（外键）
	Field        string //附件查询字段
	Arg          string //filter 参数
}

func ApiSearchJoinWith[T any](joins []*Join, withs []*With, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//先取出join字段
		joinValues := map[string]any{}
		for _, j := range joins {
			joinValues[j.Arg] = body.Filter[j.Arg]
			delete(body.Filter, j.Arg)
		}

		table := db.Engine().TableName(new(T))
		query := body.ToJoinQuery(table)

		var s []string
		//查询字段
		fs := ctx.QueryArray("field")
		if len(fs) > 0 {
			for _, f := range fs {
				s = append(s, table+"."+f)
			}
		} else if len(fields) > 0 {
			for _, f := range fields {
				s = append(s, table+"."+f)
			}
		} else {
			s = append(s, table+".*")
		}

		var datum []*T
		//var datum []map[string]any
		//session := query.Table(table)

		//补充字段
		for i, j := range withs {
			name := "w" + string(rune('a'+i))
			s = append(s, name+"."+db.Engine().Quote(j.Field)+" as "+db.Engine().Quote(j.As))
		}
		query.Select(strings.Join(s, ","))

		//连接查询
		for i, j := range joins {
			name := "j" + string(rune('a'+i))
			query.Join("INNER", []string{j.Table, name},
				name+"."+db.Engine().Quote(j.ForeignField)+"="+
					table+"."+db.Engine().Quote(j.LocaleField)).
				And(db.Engine().Quote(j.Field)+"=?", joinValues[j.Arg])
		}

		//左外连接查询
		for i, j := range withs {
			name := "w" + string(rune('a'+i))
			query.Join("LEFT OUTER", []string{j.Table, name},
				name+"."+db.Engine().Quote(j.ForeignField)+"="+
					table+"."+db.Engine().Quote(j.LocaleField))
		}

		cnt, err := query.FindAndCount(&datum)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//OK(ctx, cs)
		api.List(ctx, datum, cnt)
	}
}
