package curd

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
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
