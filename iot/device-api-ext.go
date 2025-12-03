package iot

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/smart"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/protocol"
)

func init() {
	api.Register("GET", "iot/device/:id/values", deviceValues)
	api.Register("GET", "iot/device/:id/sync", deviceSync)
	api.Register("GET", "iot/device/:id/read", deviceRead)
	api.Register("POST", "iot/device/:id/write", deviceWrite)

	api.Register("GET", "iot/device/extend/fields", deviceExtendFields)
}

func deviceExtendFields(ctx *gin.Context) {
	var fields []*smart.Field
	protocols.Range(func(name string, item *protocol.Protocol) bool {
		fields = append(fields, item.DeviceExtendFields...)
		return true
	})
	api.OK(ctx, fields)
}

func deviceValues(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "设备未上线")
		return
	}
	api.OK(ctx, d.values.Get())
}

func deviceSync(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "设备未上线")
		return
	}

	values, err := d.Sync(60)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, values)
}

func deviceRead(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "设备未上线")
		return
	}

	points := ctx.QueryArray("point")
	values, err := d.Read(points, 30)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, values)
}

func deviceWrite(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "设备未上线")
		return
	}

	var values map[string]any
	err := ctx.ShouldBind(&values)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	result, err := d.Write(values, 30)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, result)
}
