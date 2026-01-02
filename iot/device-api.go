package iot

import (
	"fmt"
	"time"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/gin-gonic/gin"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

func init() {
	//物模型
	api.Register("GET", "iot/device/:id/model", curd.ApiGet[DeviceModel]())
	api.Register("POST", "iot/device/:id/model", deviceModelUpdate)

	//执行操作
	api.Register("POST", "iot/device/:id/action/:action", deviceAction)

	//参数
	api.Register("GET", "iot/device/:id/setting/:name", deviceSetting)
	api.Register("POST", "iot/device/:id/setting/:name", deviceSettingUpdate)
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
	Version int            `json:"version,omitempty" xorm:"version"`
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

	has, err := db.Engine().ID(schemas.PK{id, name}).Get(&setting)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//修改为新内容
	setting.Content = content
	if !has {
		setting.Id = id
		setting.Name = name
		_, err = db.Engine().Insert(&setting)
	} else {
		_, err = db.Engine().ID(schemas.PK{id, name}).Cols("content").Update(&setting)
	}
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//下发最新配置
	go func() {
		var setting DeviceSetting
		has, err := db.Engine().ID(schemas.PK{id, name}).Get(&setting)
		if err != nil {
			log.Error(err)
			return
		}
		if has {
			//下发到设备
			topic := fmt.Sprintf("device/%s/setting", id)
			mqtt.Publish(topic, &setting)
		}
	}()

	api.OK(ctx, nil)
}

func deviceAction(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil {
		api.Fail(ctx, "设备未上线")
		return
	}
	action := ctx.Param("action")

	var values map[string]any
	err := ctx.ShouldBind(&values)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	result, err := d.Action(action, values, 30)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, result)
}

func deviceNear(ctx *gin.Context) {
	var devices []Device
	err := db.Engine().Where(builder.Like{"geo_code", ctx.Param("geo_code") + "%"}).Limit(1000).Find(&devices)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, devices)
}
