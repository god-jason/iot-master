package iot

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/product"
)

func init() {

	//物模型
	api.Register("GET", "iot/product/:id/model", productModel)
	api.Register("POST", "iot/product/:id/model", productModelUpdate)
}

func productModel(ctx *gin.Context) {
	id := ctx.Param("id")
	var model product.ProductModel

	_, err := db.Engine().ID(id).Get(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}

func productModelUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	var model product.ProductModel
	err := ctx.ShouldBind(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	model.Id = id

	_, err = db.Engine().ID(id).Delete(new(product.ProductModel)) //不管有没有都删掉
	_, err = db.Engine().ID(id).Insert(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}
