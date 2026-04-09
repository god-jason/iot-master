package iot

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/product"
)

func init() {

	//扩展字段
	api.Register("GET", "device/extend/fields", deviceExtendFields)
	api.Register("GET", "device/:id/extend/fields", deviceExtendFields)

	//设备绑定(当前租户)
	api.Register("GET", "device/:id/bind", deviceBind)

	//设备绑定(组织)
	api.Register("GET", "device/:id/bind/:gid", deviceBindGroup)
	api.Register("GET", "device/:id/unbind", deviceUnbindGroup)
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
	tenantId := ctx.GetString("tenant")

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

	if dev.TenantId != "" {
		api.Fail(ctx, "设备已经被绑定")
		return
	}

	if dev.GatewayId != "" {
		api.Fail(ctx, "子设备不能被绑定")
		return
	}

	var dev2 Device
	dev2.TenantId = tenantId
	_, err = db.Engine().ID(id).Cols("tenant_id").Update(&dev2)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, "")
}

func deviceBindGroup(ctx *gin.Context) {
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

	if dev.GatewayId != "" {
		api.Fail(ctx, "子设备不能被绑定")
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

func deviceUnbindGroup(ctx *gin.Context) {
	id := ctx.Param("id")

	var dev2 Device
	_, err := db.Engine().ID(id).Cols("group_id").Update(&dev2)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, "")
}
