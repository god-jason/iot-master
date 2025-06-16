package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/product"
	"xorm.io/xorm/schemas"
)

func init() {
	api.Register("GET", "iot/product/list", curd.ApiList[product.Product]())
	api.Register("POST", "iot/product/search", curd.ApiSearch[product.Product]())
	api.Register("POST", "iot/product/create", curd.ApiCreate[product.Product]())
	api.Register("GET", "iot/product/:id", curd.ApiGet[product.Product]())
	api.Register("POST", "iot/product/:id", curd.ApiUpdate[product.Product]("id", "name", "description", "type", "version", "protocol", "disabled"))
	api.Register("GET", "iot/product/:id/delete", curd.ApiDelete[product.Product]())
	api.Register("GET", "iot/product/:id/enable", curd.ApiDisable[product.Product](false))
	api.Register("GET", "iot/product/:id/disable", curd.ApiDisable[product.Product](true))

	//物模型
	api.Register("GET", "iot/product/:id/model", curd.ApiGet[product.Model]())
	api.Register("POST", "iot/product/:id/model", productModelUpdate)

	//配置接口，一般用于协议点表等
	api.Register("GET", "iot/product/:id/config/:name", productConfig)
	api.Register("POST", "iot/product/:id/config/:name", productConfigUpdate)
}

func productModelUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	var model product.Model
	err := ctx.ShouldBind(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	model.Id = id

	_, err = db.Engine().ID(id).Delete(new(product.Model)) //不管有没有都删掉
	_, err = db.Engine().ID(id).Insert(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}

func productConfig(ctx *gin.Context) {
	var config product.ProductConfig
	has, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("name")}).Get(&config)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "找不到配置文件")
		return
	}

	api.OK(ctx, config.Content)
}

func productConfigUpdate(ctx *gin.Context) {
	//body, err := io.ReadAll(ctx.Request.Body)
	//if err != nil {
	//	api.Error(ctx, err)
	//	return
	//}
	//
	var body map[string]any
	err := ctx.ShouldBind(&body)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	config := product.ProductConfig{
		Id:      ctx.Param("id"),
		Name:    ctx.Param("name"),
		Content: body,
	}

	_, err = db.Engine().ID(schemas.PK{config.Id, config.Name}).Delete(new(product.ProductConfig))
	_, err = db.Engine().ID(schemas.PK{config.Id, config.Name}).Insert(&config)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &config)
}
