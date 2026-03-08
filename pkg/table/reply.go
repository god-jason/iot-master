package table

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReplyData[T any] struct {
	Data  T      `json:"data"`
	Error string `json:"error,omitempty"`
}

type ReplyList[T any] struct {
	Data  []T    `json:"data"`
	Total int64  `json:"total"`
	Error string `json:"error,omitempty"`
}

func List(ctx *gin.Context, data any, total int64) {
	ctx.JSON(http.StatusOK, gin.H{"data": data, "total": total})
}

func OK(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, gin.H{"data": data})
}

func Fail(ctx *gin.Context, err string) {
	ctx.JSON(http.StatusOK, gin.H{"error": err})
}

func Error(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
}
