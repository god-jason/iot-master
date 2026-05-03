package iot

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/spf13/viper"
	"xorm.io/xorm/schemas"
)

func init() {
	//参数
	api.Register("GET", "product/:id/setting/:name", productSetting)
	api.Register("POST", "product/:id/setting/:name", productSettingUpdate)
}

type ProductSetting struct {
	Id       string           `json:"id" xorm:"pk"`
	Name     string           `json:"name" xorm:"pk"`
	TenantId string           `json:"tenant_id,omitempty" xorm:"index"`
	Version  int              `json:"version,omitempty" xorm:"version"`
	Content  []map[string]any `json:"content,omitempty" xorm:"json"`
	Created  time.Time        `json:"created,omitempty" xorm:"created"`
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

	//禁止租户账号修改公共产品库
	if viper.GetBool("tenant") {
		tid := ctx.GetString("tenant")
		if tid != "" {
			if tid != setting.TenantId {
				api.Fail(ctx, "禁止修改")
				return
			}
		}
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
