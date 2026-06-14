package table

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ApiCount 统计数量
func ApiCount(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var body ParamSearch
	err = ctx.ShouldBindJSON(&body)
	if err != nil {
		Error(ctx, err)
		return
	}

	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			column := table.Column("tenant_id")
			if column != nil {
				if body.Filter == nil {
					body.Filter = make(map[string]any)
				}
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