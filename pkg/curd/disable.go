package curd

import (
	"reflect"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/gin-gonic/gin"
)

func ApiDisable[T any](disable bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//id := ctx.GetInt64("id")
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//value := reflect.New(mod)
		//value.Elem().FieldByName("Disabled").SetBool(disable)
		//data := value.Interface()
		var data T
		value := reflect.ValueOf(&data).Elem()
		field := value.FieldByName("Disabled")
		field.SetBool(disable)

		_, err = db.Engine().ID(id).Cols("disabled").Update(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, nil)
	}
}

func ApiDisableHook[T any](disable bool, before, after func(id any) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//id := ctx.GetInt64("id")
		id, err := GetId(ctx)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		if before != nil {
			if err := before(id); err != nil {
				api.Error(ctx, err)
				return
			}
		}

		//value := reflect.New(mod)
		//value.Elem().FieldByName("Disabled").SetBool(disable)
		//data := value.Interface()
		var data T
		value := reflect.ValueOf(&data).Elem()
		field := value.FieldByName("Disabled")
		field.SetBool(disable)

		_, err = db.Engine().ID(id).Cols("disabled").Update(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		if after != nil {
			//改为异常执行，减少前端错误
			go func() {
				if err := after(id); err != nil {
					log.Error(err)
				}
			}()
		}

		api.OK(ctx, nil)
	}
}
