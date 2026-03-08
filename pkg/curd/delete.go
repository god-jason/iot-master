package curd

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/gin-gonic/gin"
)

func ApiDelete[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		var data T
		_, err = db.Engine().ID(id).Delete(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, nil)
	}
}

func ApiDeleteHook[T any](before, after func(m *T) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		var data T
		has, err := db.Engine().ID(id).Get(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}
		if !has {
			api.Fail(ctx, "找不到记录")
			return
		}

		if before != nil {
			if err := before(&data); err != nil {
				api.Error(ctx, err)
				return
			}
		}

		_, err = db.Engine().ID(id).Delete(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		if after != nil {
			//改为异常执行，减少前端错误
			go func() {
				if err := after(&data); err != nil {
					log.Error(err)
				}
			}()
		}

		api.OK(ctx, nil)
	}
}
