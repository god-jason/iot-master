package iot

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
)

func init() {
	api.Register("GET", "iot/link/:id/close", func(ctx *gin.Context) {
		//TODO 发布消息给linker，关闭对应连接
	})

}
