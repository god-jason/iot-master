package curd

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/db"
	"strings"
)

type With struct {
	Table        string //表名
	LocaleField  string //主表字段
	ForeignField string //附表字段（外键）
	Field        string //取字段 TODO 可以改为数组
	As           string //赋值
}

func ApiSearchWith[T any](withs []*With, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			Error(ctx, err)
			return
		}

		table := db.Engine.TableName(new(T))
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
			name := string(rune('a' + i))
			s = append(s, name+"."+db.Engine.Quote(j.Field)+" as "+db.Engine.Quote(j.As))
		}
		query.Select(strings.Join(s, ","))

		//连接查询
		for i, j := range withs {
			name := string(rune('a' + i))
			query.Join("LEFT OUTER", []string{j.Table, name},
				name+"."+db.Engine.Quote(j.ForeignField)+"="+
					table+"."+db.Engine.Quote(j.LocaleField))
		}

		cnt, err := query.FindAndCount(&datum)
		if err != nil {
			Error(ctx, err)
			return
		}

		//OK(ctx, cs)
		List(ctx, datum, cnt)
	}
}

func ApiSearchWithHook[T any](join []*With, after func(datum []*T) error, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			Error(ctx, err)
			return
		}

		table := db.Engine.TableName(new(T))
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
		for i, j := range join {
			name := string(rune('a' + i))
			s = append(s, name+"."+db.Engine.Quote(j.Field)+" as "+db.Engine.Quote(j.As))
		}
		query.Select(strings.Join(s, ","))

		//连接查询
		for i, j := range join {
			name := string(rune('a' + i))
			query.Join("LEFT OUTER", []string{j.Table, name},
				name+"."+db.Engine.Quote(j.ForeignField)+"="+
					table+"."+db.Engine.Quote(j.LocaleField))
		}

		cnt, err := query.FindAndCount(&datum)
		if err != nil {
			Error(ctx, err)
			return
		}

		//后续处理
		if after != nil {
			if err := after(datum); err != nil {
				Error(ctx, err)
				return
			}
		}

		//OK(ctx, cs)
		List(ctx, datum, cnt)
	}
}
