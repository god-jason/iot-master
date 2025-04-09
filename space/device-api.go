package space

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"xorm.io/xorm/schemas"
)

func init() {
	api.Register("GET", "iot/space/:id/device/list", spaceDeviceList)
	api.Register("GET", "iot/space/:id/device/:device/bind", spaceDeviceBind)
	api.Register("GET", "iot/space/:id/device/:device/unbind", spaceDeviceUnbind)
	api.Register("POST", "iot/space/:id/device/:device", spaceDeviceUpdate)
}

// @Summary 空间设备列表
// @Schemes
// @Description 空间设备列表
// @Tags space-device
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]SpaceDevice] 返回空间设备信息
// @Router iot/space/{id}/device/list [get]
func spaceDeviceList(ctx *gin.Context) {
	var pds []SpaceDevice
	err := db.Engine().
		Select("space_device.space_id, space_device.device_id, space_device.created, device.name as device").
		Join("INNER", "device", "device.id=space_device.device_id").
		Where("space_device.space_id=?", ctx.Param("id")).
		Find(&pds)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, pds)
}

// @Summary 绑定空间设备
// @Schemes
// @Description 绑定空间设备
// @Tags space-device
// @Param id path int true "项目ID"
// @Param device path int true "设备ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/space/{id}/device/{device}/bind [get]
func spaceDeviceBind(ctx *gin.Context) {
	pd := SpaceDevice{
		SpaceId:  ctx.Param("id"),
		DeviceId: ctx.Param("device"),
	}
	_, err := db.Engine().InsertOne(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 删除空间设备
// @Schemes
// @Description 删除空间设备
// @Tags space-device
// @Param id path int true "项目ID"
// @Param device path int true "设备ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/space/{id}/device/{device}/unbind [get]
func spaceDeviceUnbind(ctx *gin.Context) {
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("device")}).Delete(new(SpaceDevice))
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 修改空间设备
// @Schemes
// @Description 修改空间设备
// @Tags space-device
// @Param id path int true "项目ID"
// @Param device path int true "设备ID"
// @Param space-device body SpaceDevice true "空间设备信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/space/{id}/device/{device} [post]
func spaceDeviceUpdate(ctx *gin.Context) {
	var pd SpaceDevice
	err := ctx.ShouldBindJSON(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	_, err = db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("device")}).
		Cols("device_id", "name", "disabled").
		Update(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}
