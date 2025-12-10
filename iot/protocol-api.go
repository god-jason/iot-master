package iot

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func init() {

	api.Register("GET", "iot/protocol/list", func(ctx *gin.Context) {
		ps, err := GetProtocols()
		if err != nil {
			api.Error(ctx, err)
			return
		}
		api.OK(ctx, ps)
	})

	api.Register("GET", "iot/protocol/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		p, err := GetProtocol(name)
		if err != nil {
			api.Error(ctx, err)
			return
		}

		api.OK(ctx, p)
		//ctx.JSON(http.StatusOK, p) //TODO 先兼容原始前端，后续再修改
	})

}
