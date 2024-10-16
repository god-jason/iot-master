package curd

import "github.com/gin-gonic/gin"

func ApiListById[T any](field string, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamList
		err := ctx.ShouldBindQuery(&body)
		if err != nil {
			Error(ctx, err)
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

		//添加条件
		id := ctx.MustGet("id")
		query.Where(field+"=?", id)

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

func ApiList[T any](fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamList
		err := ctx.ShouldBindQuery(&body)
		if err != nil {
			Error(ctx, err)
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

func ApiListHook[T any](after func(datum []*T) error, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamList
		err := ctx.ShouldBindQuery(&body)
		if err != nil {
			Error(ctx, err)
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

		var datum []*T
		cnt, err := query.FindAndCount(&datum)
		if err != nil {
			Error(ctx, err)
			return
		}

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

func ApiListMapHook[T any](after func(datum []map[string]any) error, fields ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body ParamList
		err := ctx.ShouldBindQuery(&body)
		if err != nil {
			Error(ctx, err)
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

		var data T
		var datum []map[string]any
		cnt, err := query.Table(data).FindAndCount(&datum)
		if err != nil {
			Error(ctx, err)
			return
		}

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
