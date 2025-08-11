package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
)

func init() {
	//物模型
	api.Register("GET", "iot/device/:id/model", curd.ApiGet[DeviceModel]())
	api.Register("POST", "iot/device/:id/model", deviceModelUpdate)

}

func getDevicesInfo(ds []*Device) error {
	for _, d := range ds {
		_ = getDeviceInfo(d)
	}
	return nil
}

func getDeviceInfo(d *Device) error {
	l := devices.Load(d.Id)
	if l != nil {
		d.Status = l.Status
	}
	return nil
}

func deviceModelUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	var model DeviceModel
	err := ctx.ShouldBind(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	model.Id = id

	_, err = db.Engine().ID(id).Delete(new(DeviceModel)) //不管有没有都删掉
	_, err = db.Engine().ID(id).Insert(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}
