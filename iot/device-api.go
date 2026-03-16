package iot

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/table"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

func init() {

	//远程操作
	api.Register("GET", "device/:id/values", deviceValues)
	api.Register("GET", "device/:id/sync", deviceSync)
	api.Register("GET", "device/:id/read", deviceRead)
	api.Register("POST", "device/:id/write", deviceWrite)
	api.Register("POST", "device/:id/action/:action", deviceAction)

	//参数
	api.Register("GET", "device/:id/setting/:name", deviceSetting)
	api.Register("POST", "device/:id/setting/:name", deviceSettingUpdate)
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

func deviceAction(ctx *gin.Context) {
	id := ctx.Param("id")

	d := devices.Load(id)
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

	//记录操作员
	tab, _ := table.Get("device_log")
	if tab != nil {
		_, _ = tab.Insert(map[string]interface{}{
			"user_id":   ctx.GetString("user"), //操作用户ID
			"device_id": id,
			"content":   "远程操作：" + action,
		})
	}

	//执行操作
	result, err := d.Action(action, values, 30)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, result)
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

	//查询最新配置，主要是版本号
	var setting2 DeviceSetting
	has, err = db.Engine().ID(schemas.PK{id, name}).Get(&setting2)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//如果设备在线，则直接通过MQTT下发配置
	dev := devices.Load(id)
	if dev != nil {
		_, err = dev.Setting(setting2.Name, setting2.Content, setting2.Version, 30)
		if err != nil {
			api.Error(ctx, err)
			return
		}
	}

	api.OK(ctx, nil)
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
