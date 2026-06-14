package table

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ApiGroup 聚合查询
func ApiGroup(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var body ParamGroup
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

	results, err := table.Group(&body)
	if err != nil {
		Error(ctx, err)
		return
	}
	OK(ctx, results)
}