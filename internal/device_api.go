package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/device"
)

func init() {
	api.Register("GET", "iot/device/list", curd.ApiListHook[Device](getDevicesInfo))
	api.Register("POST", "iot/device/search", curd.ApiSearchHook[Device](getDevicesInfo))
	api.Register("POST", "iot/device/create", curd.ApiCreateHook[Device](nil, func(m *Device) error {
		//TODO 加载设备
		return nil
	}))
	api.Register("GET", "iot/device/:id", curd.ApiGetHook[Device](getDeviceInfo))
	api.Register("POST", "iot/device/:id", curd.ApiUpdateHook[Device](nil, func(m *Device) error {
		//TODO 重新加载设备
		return nil
	}, "id", "name", "description", "product_id", "link_id", "disabled", "station"))

	api.Register("GET", "iot/device/:id/delete", curd.ApiDeleteHook[Device](nil, func(m *Device) error {
		return UnloadDevice(m.Id)
	}))

	api.Register("GET", "iot/device/:id/enable", curd.ApiDisableHook[Device](false, nil, func(id any) error {
		_, err := LoadDevice(id.(string))
		return err
	}))

	api.Register("GET", "iot/device/:id/disable", curd.ApiDisableHook[Device](true, nil, func(id any) error {
		return UnloadDevice(id.(string))
	}))

	//物模型
	api.Register("GET", "iot/device/:id/model", curd.ApiGet[device.DeviceModel]())
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

	var model device.DeviceModel
	err := ctx.ShouldBind(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	model.Id = id

	_, err = db.Engine().ID(id).Delete(new(device.DeviceModel)) //不管有没有都删掉
	_, err = db.Engine().ID(id).Insert(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}
