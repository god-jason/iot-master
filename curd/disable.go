package curd

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/db"
	"reflect"
)

func ApiDisable[T any](disable bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//id := ctx.GetInt64("id")
		id := ctx.MustGet("id")

		//value := reflect.New(mod)
		//value.Elem().FieldByName("Disabled").SetBool(disable)
		//data := value.Interface()
		var data T
		value := reflect.ValueOf(&data).Elem()
		field := value.FieldByName("Disabled")
		field.SetBool(disable)

		_, err := db.Engine.ID(id).Cols("disabled").Update(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		OK(ctx, nil)
	}
}

func ApiDisableHook[T any](disable bool, before, after func(id any) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//id := ctx.GetInt64("id")
		id := ctx.MustGet("id")
		if before != nil {
			if err := before(id); err != nil {
				Error(ctx, err)
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

		_, err := db.Engine.ID(id).Cols("disabled").Update(&data)
		if err != nil {
			Error(ctx, err)
			return
		}

		if after != nil {
			if err := after(id); err != nil {
				Error(ctx, err)
				return
			}
		}

		OK(ctx, nil)
	}
}
