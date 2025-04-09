package product

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"xorm.io/xorm/schemas"
)

func init() {
	api.Register("GET", "iot/product/list", curd.ApiList[Product]())
	api.Register("POST", "iot/product/search", curd.ApiSearch[Product]())
	api.Register("POST", "iot/product/create", curd.ApiCreate[Product]())
	api.Register("GET", "iot/product/:id", curd.ApiGet[Product]())
	api.Register("POST", "iot/product/:id", curd.ApiUpdate[Product]("id", "name", "description", "type", "version", "protocol", "disabled"))
	api.Register("GET", "iot/product/:id/delete", curd.ApiDelete[Product]())
	api.Register("GET", "iot/product/:id/enable", curd.ApiDisable[Product](false))
	api.Register("GET", "iot/product/:id/disable", curd.ApiDisable[Product](true))

	//物模型
	api.Register("GET", "iot/product/:id/model", curd.ApiGet[ProductModel]())
	api.Register("POST", "iot/product/:id/model", productModelUpdate)

	//配置接口，一般用于协议点表等
	api.Register("GET", "iot/product/:id/config/:name", productConfig)
	api.Register("POST", "iot/product/:id/config/:name", productConfigUpdate)
}

func productModelUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	var model ProductModel
	err := ctx.ShouldBind(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	model.Id = id

	_, err = db.Engine().ID(id).Delete(new(ProductModel)) //不管有没有都删掉
	_, err = db.Engine().ID(id).Insert(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}

func productConfig(ctx *gin.Context) {
	var config ProductConfig
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

	config := ProductConfig{
		Id:      ctx.Param("id"),
		Name:    ctx.Param("name"),
		Content: body,
	}

	_, err = db.Engine().ID(schemas.PK{config.Id, config.Name}).Delete(new(ProductConfig))
	_, err = db.Engine().ID(schemas.PK{config.Id, config.Name}).Insert(&config)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &config)
}
