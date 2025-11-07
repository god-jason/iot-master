package iot

import (
	"time"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"xorm.io/xorm/schemas"
)

func init() {
	//物模型
	api.Register("GET", "iot/device/:id/model", curd.ApiGet[DeviceModel]())
	api.Register("POST", "iot/device/:id/model", deviceModelUpdate)

	//参数
	api.Register("GET", "iot/device/:id/setting/:name", deviceSetting)
	api.Register("POST", "iot/device/:id/setting/:name", deviceSettingUpdate)

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

type DeviceSetting struct {
	Id      string         `json:"id" xorm:"pk"`
	Name    string         `json:"name" xorm:"pk"`
	Content map[string]any `json:"content,omitempty" xorm:"text"`
	Created time.Time      `json:"created,omitempty" xorm:"created"`
}

func deviceSetting(ctx *gin.Context) {
	id := ctx.Param("id")
	name := ctx.Param("name")

	var setting DeviceSetting
	has, err := db.Engine().ID(schemas.PK{id, name}).Get(&setting)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.OK(ctx, nil)
		return
	}

	api.OK(ctx, &setting.Content)
}

func deviceSettingUpdate(ctx *gin.Context) {
	id := ctx.Param("id")
	name := ctx.Param("name")

	var content map[string]any
	err := ctx.ShouldBind(&content)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	var setting DeviceSetting
	setting.Id = id
	setting.Name = name
	setting.Content = content

	_, err = db.Engine().ID(schemas.PK{id, name}).Delete(new(DeviceSetting)) //不管有没有都删掉
	_, err = db.Engine().ID(schemas.PK{id, name}).Insert(&setting)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, content)
}
