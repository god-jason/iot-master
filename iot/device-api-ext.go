package iot

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/product"
)

func init() {
	api.Register("GET", "iot/device/:id/values", deviceValues)
	api.Register("GET", "iot/device/:id/sync", deviceSync)
	api.Register("GET", "iot/device/:id/read", deviceRead)
	api.Register("POST", "iot/device/:id/write", deviceWrite)

	api.Register("GET", "iot/device/extend/fields", deviceExtendFields)
	api.Register("GET", "iot/device/:id/extend/fields", deviceExtendFields)

	api.Register("GET", "iot/device/:id/bind/:gid", deviceBind)
	api.Register("GET", "iot/device/:id/unbind", deviceUnbind)
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

func deviceExtendFields(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		api.OK(ctx, _deviceExtendFields)
		return
	}

	var dev Device
	has, err := db.Engine().ID(id).Get(&dev)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "device not found")
		return
	}

	if dev.ProductId == "" {
		api.OK(ctx, nil)
		return
	}

	var prod product.Product
	has, err = db.Engine().ID(dev.ProductId).Get(&prod)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "product not found")
		return
	}

	if prod.Protocol == "" {
		api.OK(ctx, nil)
		return
	}

	proto, err := GetProtocol(prod.Protocol)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, proto.DeviceExtendFields)
}

func deviceBind(ctx *gin.Context) {
	id := ctx.Param("id")
	gid := ctx.Param("gid")

	var dev Device
	has, err := db.Engine().ID(id).Get(&dev)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "设备不存在")
		return
	}

	if dev.GroupId != "" {
		api.Fail(ctx, "设备已经被绑定")
		return
	}

	var dev2 Device
	dev2.GroupId = gid
	_, err = db.Engine().ID(id).Cols("group_id").Update(&dev2)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, "")
}

func deviceUnbind(ctx *gin.Context) {
	id := ctx.Param("id")

	var dev2 Device
	_, err := db.Engine().ID(id).Cols("group_id").Update(&dev2)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, "")
}
