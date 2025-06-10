package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "iot/device/:id/values", deviceValues)
	api.Register("GET", "iot/device/:id/values/refresh", deviceValuesRefresh)
}

func deviceValues(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "设备未上线")
		return
	}
	api.OK(ctx, d.Values)
}

func deviceValuesRefresh(ctx *gin.Context) {
	api.OK(ctx, "暂未支持")
}
