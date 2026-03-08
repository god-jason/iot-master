package curd

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Noop(ctx *gin.Context) {
	ctx.String(http.StatusForbidden, "Unsupported")
}
