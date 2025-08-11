package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "iot/link/:id/close", func(ctx *gin.Context) {
		//TODO 发布消息给linker，关闭对应连接
	})

}
