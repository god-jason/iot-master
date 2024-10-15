package curd

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Noop(ctx *gin.Context) {
	ctx.String(http.StatusForbidden, "Unsupported")
}
