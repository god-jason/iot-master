package curd

import (
	"reflect"

	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func ApiCount[T any]() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var body ParamSearch
		err := ctx.ShouldBindJSON(&body)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		var d T
		//多租户处理
		tid := ctx.GetString("tenant")
		if tid != "" {
			//只有未传值tenant_id时，才会赋值用户所在的tenant_id
			if _, ok := body.Filter["tenant_id"]; !ok {
				field := reflect.ValueOf(&d).Elem().FieldByName("TenantId")
				if field.IsValid() && field.IsZero() && field.Kind() == reflect.String {
					body.Filter["tenant_id"] = tid
				}
			}
		}

		query := body.ToQuery()

		cnt, err := query.Count(d)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, cnt)
	}
}
