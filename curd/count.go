package curd

import (
	"github.com/gin-gonic/gin"
)

func ApiCount[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			Error(ctx, err)
			return
		}

		query := body.ToQuery()

		var d T
		cnt, err := query.Count(d)
		if err != nil {
			Error(ctx, err)
			return
		}

		OK(ctx, cnt)
	}
}
