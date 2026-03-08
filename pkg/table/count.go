package table

import (
	"slices"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/smart"
	"github.com/gin-gonic/gin"
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
	tid := ctx.GetString("tenant")
	if tid != "" {
		tenantId := slices.IndexFunc(table.Columns, func(column *smart.Column) bool {
			return column.Name == "tenant_id"
		})
		if tenantId > -1 {
			//只有未传值tenant_id时，才会赋值用户所在的tenant_id
			if _, ok := body.Filter["tenant_id"]; !ok {
				body.Filter["tenant_id"] = tid
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
