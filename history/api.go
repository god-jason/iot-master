package history

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
)

func init() {
	api.Register("GET", "device/:id/history/:point", deviceHistory)
	//api.Register("GET", "influxdb/device/:id/history/:point", deviceHistory)
}

type Device struct {
	Id        string `json:"id" xorm:"pk"`
	ProductId string `json:"product_id" xorm:"index"`
}

func deviceHistory(ctx *gin.Context) {
	var dev Device
	has, err := db.Engine().ID(ctx.Param("id")).Get(&dev)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "找不到设备")
		return
	}

	key := ctx.Param("point")
	start := ctx.DefaultQuery("start", "-5h")
	end := ctx.DefaultQuery("end", "0h")
	window := ctx.DefaultQuery("window", "10m")
	method := ctx.DefaultQuery("method", "mean") //last

	points, err := Query(dev.ProductId, dev.Id, key, start, end, window, method)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, points)
}
