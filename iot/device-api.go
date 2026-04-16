package iot

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/mqtt"
	"github.com/god-jason/iot-master/pkg/table"
	"github.com/spf13/cast"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

func init() {

	//状态
	api.Register("GET", "device/:id/values", deviceValues)

	//远程操作
	api.Register("GET", "device/:id/sync", deviceSync)
	api.Register("GET", "device/:id/read", deviceRead)
	api.Register("POST", "device/:id/write", deviceWrite)
	api.Register("POST", "device/:id/action/:action", deviceAction)
	api.Register("GET", "device/:id/error/clear", deviceErrorClear)

	//子设备操作
	api.Register("GET", "device/:id/sync/:child", deviceSync)
	api.Register("GET", "device/:id/read/:child", deviceRead)
	api.Register("POST", "device/:id/write/:child", deviceWrite)

	//参数操作
	api.Register("GET", "device/:id/setting/:name", deviceSetting)
	api.Register("POST", "device/:id/setting/:name", deviceSettingUpdate)
	api.Register("GET", "device/:id/setting/clear", deviceSettingClear) //清空云端配置

	//同步数据库
	api.Register("GET", "device/:id/download/:database", deviceDownloadDatabase)
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
	if d == nil || !d.Online {
		api.Fail(ctx, "设备未上线")
		return
	}

	values, err := d.Sync(60, ctx.Param("child"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, values)
}

func deviceRead(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil || !d.Online {
		api.Fail(ctx, "设备未上线")
		return
	}

	points := ctx.QueryArray("point")
	values, err := d.Read(points, 30, ctx.Param("child"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, values)
}

func deviceWrite(ctx *gin.Context) {
	d := devices.Load(ctx.Param("id"))
	if d == nil || !d.Online {
		api.Fail(ctx, "设备未上线")
		return
	}

	var values map[string]any
	err := ctx.ShouldBind(&values)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	result, err := d.Write(values, 30, ctx.Param("child"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, result)
}

func deviceAction(ctx *gin.Context) {
	id := ctx.Param("id")

	d := devices.Load(id)
	if d == nil || !d.Online {
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
		//TODO 获取名称
		content := "远程操作："

		//操作名
		if val, ok := values["name"]; ok {
			content += cast.ToString(val)
		} else {
			content += action
		}

		//操作值
		if val, ok := values["value"]; ok {
			if has, ok := val.(bool); ok {
				if has {
					content += " 打开"
				} else {
					content += " 关闭"
				}
			}
		}

		//记录日志
		_, _ = tab.Insert(map[string]interface{}{
			"user_id":   ctx.GetString("user"), //操作用户ID
			"device_id": id,
			"content":   content,
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

func deviceErrorClear(ctx *gin.Context) {
	id := ctx.Param("id")

	var d Device
	_, err := db.Engine().ID(id).Cols("error", "error_string").Update(&d)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	// 如果在线，则直接清理错误
	dev := devices.Load(id)
	if dev != nil {
		// 清空错误状态
		dev.PutValues(map[string]any{
			"error":        false,
			"error_string": "",
		})
		//_, err = dev.Action("clear_error", nil, 30)
		//if err != nil {
		//	api.Error(ctx, err)
		//	return
		//}
	}

	api.OK(ctx, nil)
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

func deviceSettingClear(ctx *gin.Context) {
	id := ctx.Param("id")

	var setting DeviceSetting
	cnt, err := db.Engine().Where("id=?", id).Delete(&setting)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, cnt)
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

	//记录操作员
	tab, _ := table.Get("device_log")
	if tab != nil {
		_, _ = tab.Insert(map[string]interface{}{
			"user_id":   ctx.GetString("user"), //操作用户ID
			"device_id": id,
			"content":   "修改配置：" + name,
		})
	}

	//如果设备在线，则直接通过MQTT下发配置
	dev := devices.Load(id)
	if dev != nil && dev.Online {
		_, err = dev.Setting(setting2.Name, setting2.Content, setting2.Version, 30)
		if err != nil {
			api.Error(ctx, err)
			return
		}
	} else {
		api.Fail(ctx, "设备未在线，配置已经保存")
		return
	}

	api.OK(ctx, nil)
}

func deviceDownloadDatabase(ctx *gin.Context) {
	id := ctx.Param("id")
	name := ctx.Param("database")

	d := devices.Load(id)
	if d == nil || !d.Online {
		api.Fail(ctx, "设备未上线，等待重新上线后会自动同步")
		return
	}

	tab, err := table.Get(name)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//查找子设备
	rows, err := tab.Find(&table.ParamSearch{
		Skip:   0,
		Limit:  999,
		Filter: map[string]any{"gateway_id": id},
	})
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//清空，并插入新数据
	if len(rows) > 0 {
		mqtt.Publish("device/"+id+"/database/"+name+"/clear", nil)
		mqtt.Publish("device/"+id+"/database/"+name+"/insertArray", rows)
	}

	api.OK(ctx, len(rows))
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
