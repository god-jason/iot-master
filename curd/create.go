package curd

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/db"
)

func ApiCreate[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data T
		err := ctx.ShouldBindJSON(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		_, err = db.Engine.InsertOne(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		OK(ctx, &data)
	}
}

func ApiCreateHook[T any](before, after func(m *T) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data T
		err := ctx.ShouldBindJSON(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		if before != nil {
			if err := before(&data); err != nil {
				Error(ctx, err)
				return
			}
		}

		_, err = db.Engine.InsertOne(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		if after != nil {
			if err := after(&data); err != nil {
				Error(ctx, err)
				return
			}
		}

		OK(ctx, &data)
	}
}
