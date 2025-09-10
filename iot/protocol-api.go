package iot

import (
	"net/http"

	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func init() {

	api.Register("GET", "iot/protocol/list", func(ctx *gin.Context) {
		api.OK(ctx, GetProtocols())
	})

	api.Register("GET", "iot/protocol/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		p := GetProtocol(name)
		if p == nil {
			api.Fail(ctx, "协议找不到")
			return
		}
		//api.OK(ctx, p)
		ctx.JSON(http.StatusOK, p) //TODO 先兼容原始前端，后续再修改
	})

}
