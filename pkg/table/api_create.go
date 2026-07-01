package table

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ApiCreate 创建数据
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

	//默认租户ID
	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			column := table.Column("tenant_id")
			if column != nil {
				if _, ok := doc["tenant_id"]; !ok {
					doc["tenant_id"] = tid
				}
			}
		}
	}

	//默认用户ID
	uc := table.Column("user_id")
	if uc != nil {
		if uid, ok := doc["user_id"]; !ok || uid == "" {
			doc["user_id"] = ctx.GetString("user")
		}
	}

	id, err := table.Insert(doc)
	if err != nil {
		Error(ctx, err)
		return
	}
	OK(ctx, id)
}
