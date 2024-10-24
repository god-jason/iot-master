package product

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/api"
	"github.com/god-jason/iot-master/curd"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
)

func init() {
	api.Register("POST", "/product/count", curd.ApiCount[Product]())
	api.Register("POST", "/product/search", curd.ApiSearch[Product]())
	api.Register("GET", "/product/list", curd.ApiList[Product]())
	api.Register("POST", "/product/create", curd.ApiCreateHook[Product](curd.GenerateID[Product](), nil))
	api.Register("GET", "/product/:id", curd.ParseParamStringId, curd.ApiGet[Product]())
	api.Register("POST", "/product/:id", curd.ParseParamStringId, curd.ApiUpdate[Product]())
	api.Register("GET", "/product/:id/delete", curd.ParseParamStringId, curd.ApiDeleteHook[Product](nil, nil))

	api.Register("GET", "/product/:id/config/:config", productConfigGet)
	api.Register("POST", "/product/:id/config/:config", productConfigSet)
}

// @Summary 查询产品数量
// @Schemes
// @Description 查询产品数量
// @Tags product
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int64] 返回产品数量
// @Router /product/count [post]
func noopProductCount() {}

// @Summary 查询产品
// @Schemes
// @Description 查询产品
// @Tags product
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Product] 返回产品信息
// @Router /product/search [post]
func noopProductSearch() {}

// @Summary 查询产品
// @Schemes
// @Description 查询产品
// @Tags product
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Product] 返回产品信息
// @Router /product/list [get]
func noopProductList() {}

// @Summary 创建产品
// @Schemes
// @Description 创建产品
// @Tags product
// @Param search body Product true "产品信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Product] 返回产品信息
// @Router /product/create [post]
func noopProductCreate() {}

// @Summary 修改产品
// @Schemes
// @Description 修改产品
// @Tags product
// @Param id path int true "产品ID"
// @Param product body Product true "产品信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Product] 返回产品信息
// @Router /product/{id} [post]
func noopProductUpdate() {}

// @Summary 获取产品
// @Schemes
// @Description 获取产品
// @Tags product
// @Param id path int true "产品ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Product] 返回产品信息
// @Router /product/{id} [get]
func noopProductGet() {}

// @Summary 删除产品
// @Schemes
// @Description 删除产品
// @Tags product
// @Param id path int true "产品ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Product] 返回产品信息
// @Router /product/{id}/delete [get]
func noopProductDelete() {}

// @Summary 获得产品配置
// @Schemes
// @Description 获得产品配置
// @Tags product
// @Param id path int true "产品ID"
// @Param config path string true "配置"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[any]
// @Router /product/{id}/config/{config} [get]
func productConfigGet(ctx *gin.Context) {
	fn := filepath.Join(viper.GetString("data"), "product", ctx.Param("id"), ctx.Param("config")+".json")
	buf, err := os.ReadFile(fn)
	if err != nil {
		//curd.Error(ctx, err)
		curd.OK(ctx, nil)
		return
	}

	var data any
	err = json.Unmarshal(buf, &data)
	if err != nil {
		curd.Error(ctx, err)
		return
	}

	curd.OK(ctx, data)
}

// @Summary 修改产品配置
// @Schemes
// @Description 修改产品配置
// @Tags product
// @Param id path int true "产品ID"
// @Param config path string true "配置"
// @Param config body any true "产品版本信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router /product/{id}/config/{config} [post]
func productConfigSet(ctx *gin.Context) {
	dir := filepath.Join(viper.GetString("data"), "product", ctx.Param("id"))
	_ = os.MkdirAll(dir, os.ModePerm)
	fn := filepath.Join(dir, ctx.Param("config")+".json")

	//清除缓存
	delete(configs, fn)

	file, err := os.Create(fn)
	if err != nil {
		curd.Error(ctx, err)
		return
	}
	defer file.Close()

	//写入文件
	_, err = io.Copy(file, ctx.Request.Body)
	if err != nil {
		curd.Error(ctx, err)
		return
	}

	curd.OK(ctx, nil)
}
