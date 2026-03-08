package curd

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func ApiSum[T any](field string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		query := body.ToQuery()

		var data T
		res, err := query.Sum(&data, field)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, res)
	}
}

func ApiGroup[T any](fun, group string, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		query := body.ToQuery()

		//查询字段
		fs := ctx.QueryArray("field")
		if len(fs) > 0 {
			query.Cols(fs...)
		} else if len(fields) > 0 {
			query.Cols(fields...)
		}
		query.Cols(fun)

		var data []*T //TODO 最后一列未存入
		err = query.GroupBy(group).Find(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, data)
	}
}

func ApiGroupDate[T any](field ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

	}
}
