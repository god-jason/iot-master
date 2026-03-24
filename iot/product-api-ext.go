package iot

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"xorm.io/xorm/schemas"
)

func init() {
	//参数
	api.Register("GET", "product/:id/setting/:name", productSetting)
	api.Register("POST", "product/:id/setting/:name", productSettingUpdate)
}

type ProductSetting struct {
	Id      string    `json:"id" xorm:"pk"`
	Name    string    `json:"name" xorm:"pk"`
	Version int       `json:"version,omitempty" xorm:"version"`
	Content any       `json:"content,omitempty" xorm:"text"`
	Created time.Time `json:"created,omitempty" xorm:"created"`
}

func productSetting(ctx *gin.Context) {
	id := ctx.Param("id")
	name := ctx.Param("name")

	var setting ProductSetting
	has, err := db.Engine().ID(schemas.PK{id, name}).Get(&setting)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.OK(ctx, nil)
		return
	}

	api.OK(ctx, &setting)
}

func productSettingUpdate(ctx *gin.Context) {
	id := ctx.Param("id")
	name := ctx.Param("name")

	var ps ProductSetting
	err := ctx.ShouldBind(&ps)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	var setting ProductSetting
	has, err := db.Engine().ID(schemas.PK{id, name}).Get(&setting)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//修改为新内容
	setting.Content = ps.Content
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

	api.OK(ctx, nil)
}
