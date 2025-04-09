package project

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"xorm.io/xorm/schemas"
)

func init() {
	api.Register("GET", "iot/project/:id/device/list", projectDeviceList)
	api.Register("GET", "iot/project/:id/device/:device/bind", projectDeviceBind)
	api.Register("GET", "iot/project/:id/device/:device/unbind", projectDeviceUnbind)
	api.Register("POST", "iot/project/:id/device/:device", projectDeviceUpdate)
}

// @Summary 空间设备列表
// @Schemes
// @Description 空间设备列表
// @Tags project-device
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]ProjectDevice] 返回空间设备信息
// @Router iot/project/{id}/device/list [get]
func projectDeviceList(ctx *gin.Context) {
	var pds []ProjectDevice
	err := db.Engine().
		Select("project_device.project_id, project_device.device_id, project_device.name, project_device.created, device.name as device").
		Join("INNER", "device", "device.id=project_device.device_id").
		Where("project_device.project_id=?", ctx.Param("id")).
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
// @Tags project-device
// @Param id path int true "项目ID"
// @Param device path int true "设备ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/device/{device}/bind [get]
func projectDeviceBind(ctx *gin.Context) {
	pd := ProjectDevice{
		ProjectId: ctx.Param("id"),
		DeviceId:  ctx.Param("device"),
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
// @Tags project-device
// @Param id path int true "项目ID"
// @Param device path int true "设备ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/device/{device}/unbind [get]
func projectDeviceUnbind(ctx *gin.Context) {
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("device")}).Delete(new(ProjectDevice))
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 修改空间设备
// @Schemes
// @Description 修改空间设备
// @Tags project-device
// @Param id path int true "项目ID"
// @Param device path int true "设备ID"
// @Param project-device body ProjectDevice true "空间设备信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/device/{device} [post]
func projectDeviceUpdate(ctx *gin.Context) {
	var pd ProjectDevice
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
