package curd

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
)

func ApiClear[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data T
		_, err := db.Engine().Where("1=1").Delete(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}
		api.OK(ctx, nil)
	}
}
