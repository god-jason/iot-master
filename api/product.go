package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/v3/model"
	"github.com/zgwit/iot-master/v3/pkg/curd"
)

// @Summary 查询产品
// @Schemes
// @Description 查询产品
// @Tags product
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[model.Product] 返回产品信息
// @Router /product/search [post]
func noopProductSearch() {}

// @Summary 查询产品
// @Schemes
// @Description 查询产品
// @Tags product
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[model.Product] 返回产品信息
// @Router /product/list [get]
func noopProductList() {}

// @Summary 创建产品
// @Schemes
// @Description 创建产品
// @Tags product
// @Param search body model.Product true "产品信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[model.Product] 返回产品信息
// @Router /product/create [post]
func noopProductCreate() {}

// @Summary 修改产品
// @Schemes
// @Description 修改产品
// @Tags product
// @Param id path int true "产品ID"
// @Param product body model.Product true "产品信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[model.Product] 返回产品信息
// @Router /product/{id} [post]
func noopProductUpdate() {}

// @Summary 删除产品
// @Schemes
// @Description 删除产品
// @Tags product
// @Param id path int true "产品ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[model.Product] 返回产品信息
// @Router /product/{id}/delete [get]
func noopProductDelete() {}

func productRouter(app *gin.RouterGroup) {

	app.POST("/search", curd.ApiSearch[model.Product]())
	app.GET("/list", curd.ApiList[model.Product]())
	app.POST("/create", curd.ApiCreate[model.Product](curd.GenerateRandomKey(8), nil))
	app.GET("/:id", curd.ParseParamStringId, curd.ApiGet[model.Product]())
	app.POST("/:id", curd.ParseParamStringId, curd.ApiModify[model.Product](nil, nil,
		"id", "name", "version", "desc", "properties", "functions", "events", "parameters", "constraints"))
	app.GET("/:id/delete", curd.ParseParamStringId, curd.ApiDelete[model.Product](nil, nil))
}