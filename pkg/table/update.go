package table

import (
	"strings"

	"github.com/gin-gonic/gin"
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
	cnt, err := table.UpdateById(id, update)
	if err != nil {
		Error(ctx, err)
		return
	}

	OK(ctx, cnt)
}
