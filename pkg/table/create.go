package table

import (
	"github.com/gin-gonic/gin"
)

func ApiCreate(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	var doc Document
	err = ctx.ShouldBindJSON(&doc)
	if err != nil {
		Error(ctx, err)
		return
	}

	//多租户创建数据，用默认租户id
	tid := ctx.GetString("tenant")
	if tid != "" {
		column := table.Column("tenant_id")
		if column != nil {
			//只有未传值tenant_id时，才会赋值用户所在的tenant_id
			if _, ok := doc["tenant_id"]; !ok {
				doc["tenant_id"] = tid
			}
		}
	}

	id, err := table.Insert(doc)
	if err != nil {
		Error(ctx, err)
		return
	}

	OK(ctx, id)
}
