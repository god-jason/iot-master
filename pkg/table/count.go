package table

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/spf13/viper"
)

func ApiCount(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var body ParamSearch
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//多租户过滤
	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			column := table.Column("tenant_id")
			if column != nil {
				if body.Filter == nil {
					body.Filter = make(map[string]any)
				}
				//只有未传值tenant_id时，才会赋值用户所在的tenant_id
				if _, ok := body.Filter["tenant_id"]; !ok {
					body.Filter["tenant_id"] = tid
				}
			}
		}
	}

	ret, err := table.Count(body.Filter)
	if err != nil {
		Error(ctx, err)
		return
	}

	OK(ctx, ret)
}
