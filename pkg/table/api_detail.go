package table

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ApiDetail 获取详情
func ApiDetail(ctx *gin.Context) {
	table, err := Get(ctx.Param("table"))
	if err != nil {
		Error(ctx, err)
		return
	}

	id := strings.TrimLeft(ctx.Param("id"), "/")
	doc, err := table.Detail(id, nil)
	if err != nil {
		Error(ctx, err)
		return
	}
	OK(ctx, doc)
}