package curd

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/db"
)

func ApiDelete[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.MustGet("id")

		var data T
		_, err := db.Engine.ID(id).Delete(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		OK(ctx, nil)
	}
}

func ApiDeleteHook[T any](before, after func(m *T) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.MustGet("id")

		var data T
		has, err := db.Engine.ID(id).Get(&data)
		if err != nil {
			Error(ctx, err)
			return
		}
		if !has {
			Fail(ctx, "找不到记录")
			return
		}

		if before != nil {
			if err := before(&data); err != nil {
				Error(ctx, err)
				return
			}
		}

		_, err = db.Engine.ID(id).Delete(&data)
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

		OK(ctx, nil)
	}
}
