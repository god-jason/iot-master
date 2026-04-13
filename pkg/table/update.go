package table

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func ApiUpdate(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var update Document
	err = ctx.ShouldBindJSON(&update)
	if err != nil {
		Error(ctx, err)
		return
	}

	id := strings.TrimLeft(ctx.Param("id"), "/")

	//多租户过滤
	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			column := table.Column("tenant_id")
			if column != nil {
				cnt, err := table.UpdateByIdEx(id, map[string]any{"tenant_id": tid}, update)
				if err != nil {
					Error(ctx, err)
					return
				}
				OK(ctx, cnt)
				return
			}
		}
	}

	cnt, err := table.UpdateById(id, update)
	if err != nil {
		Error(ctx, err)
		return
	}

	OK(ctx, cnt)
}
