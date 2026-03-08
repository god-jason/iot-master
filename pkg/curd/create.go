package curd

import (
	"reflect"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func ApiCreate[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data T
		err := ctx.ShouldBindJSON(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//默认值
		field := reflect.ValueOf(&data).Elem().FieldByName("Id")
		if field.IsValid() && field.IsZero() && field.Kind() == reflect.String {
			key := xid.New().String()
			field.SetString(key)
		}

		//多租户处理
		tid := ctx.GetString("tenant")
		if tid != "" {
			field := reflect.ValueOf(&data).Elem().FieldByName("TenantId")
			if field.IsValid() && field.IsZero() && field.Kind() == reflect.String {
				field.SetString(tid)
			}
		}

		_, err = db.Engine().InsertOne(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, &data)
	}
}

func ApiCreateHook[T any](before, after func(m *T) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data T
		err := ctx.ShouldBindJSON(&data)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		//默认值
		field := reflect.ValueOf(&data).Elem().FieldByName("Id")
		if field.IsValid() && field.IsZero() && field.Kind() == reflect.String {
			key := xid.New().String()
			field.SetString(key)
		}

		//多租户处理
		tid := ctx.GetString("tenant")
		if tid != "" {
			field := reflect.ValueOf(&data).Elem().FieldByName("TenantId")
			if field.IsValid() && field.IsZero() && field.Kind() == reflect.String {
				field.SetString(tid)
			}
		}

		if before != nil {
			if err := before(&data); err != nil {
				api.Error(ctx, err)
				return
			}
		}

		_, err = db.Engine().InsertOne(&data)
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

		api.OK(ctx, &data)
	}
}
