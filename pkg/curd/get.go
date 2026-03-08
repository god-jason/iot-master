package curd

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
)

func ApiGet[T any](fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		query := db.Engine().ID(id)
		//查询字段
		fs := ctx.QueryArray("field")
		if len(fs) > 0 {
			query.Cols(fs...)
		} else if len(fields) > 0 {
			query.Cols(fields...)
		}

		var data T
		has, err := query.Get(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		} else if !has {
			api.Fail(ctx, "记录不存在")
			return
		}
		api.OK(ctx, &data)
	}
}

func ApiGetHook[T any](after func(m *T) error, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		query := db.Engine().ID(id)
		//查询字段
		fs := ctx.QueryArray("field")
		if len(fs) > 0 {
			query.Cols(fs...)
		} else if len(fields) > 0 {
			query.Cols(fields...)
		}

		var data T
		has, err := query.Get(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		} else if !has {
			api.Fail(ctx, "记录不存在")
			return
		}

		if after != nil {
			if err := after(&data); err != nil {
				api.Error(ctx, err)
				return
			}
		}

		api.OK(ctx, &data)
	}
}

func ApiGetMapHook[T any](after func(m map[string]any) error, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		query := db.Engine().ID(id)
		//查询字段
		fs := ctx.QueryArray("field")
		if len(fs) > 0 {
			query.Cols(fs...)
		} else if len(fields) > 0 {
			query.Cols(fields...)
		}

		var data T
		var m map[string]any
		has, err := query.Table(data).Get(&m)
		if err != nil {
			api.Error(ctx, err)
			return
		} else if !has {
			api.Fail(ctx, "记录不存在")
			return
		}

		if after != nil {
			if err := after(m); err != nil {
				api.Error(ctx, err)
				return
			}
		}

		api.OK(ctx, &data)
	}
}
