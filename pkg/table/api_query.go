package table

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ApiQuery 根据ID查询单条数据
func ApiQuery(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	id := strings.TrimLeft(ctx.Param("id"), "/")
	doc, err := table.Get(id, nil)
	if err != nil {
		Error(ctx, err)
		return
	}
	OK(ctx, doc)
}