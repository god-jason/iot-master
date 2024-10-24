package curd

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/db"
	"strings"
)

func ApiListByIdWith[T any](field string, withs []*With, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamList
		err := ctx.ShouldBindQuery(&body)
		if err != nil {
			Error(ctx, err)
			return
		}

		table := db.Engine.TableName(new(T))
		query := body.ToQuery()

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

		//添加条件
		id := ctx.MustGet("id")
		query.Where(table+"."+db.Engine.Quote(field)+"=?", id)

		var datum []*T
		cnt, err := query.FindAndCount(&datum)
		if err != nil {
			Error(ctx, err)
			return
		}

		//OK(ctx, cs)
		List(ctx, datum, cnt)
	}
}

func ApiListWith[T any](withs []*With, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamList
		err := ctx.ShouldBindQuery(&body)
		if err != nil {
			Error(ctx, err)
			return
		}

		table := db.Engine.TableName(new(T))
		query := body.ToQuery()

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

		var datum []*T
		cnt, err := query.FindAndCount(&datum)
		if err != nil {
			Error(ctx, err)
			return
		}

		//OK(ctx, cs)
		List(ctx, datum, cnt)
	}
}
