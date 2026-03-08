package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/config"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func init() {
	api.Register("POST", "setting/:module", settingSet)
	api.Register("GET", "setting/:module", settingGet)
	api.Register("GET", "setting/:module/form", settingForm)
	api.Register("GET", "settings", settingModules)
}

// @Summary 查询配置
// @Schemes
// @Description 查询配置
// @Tags setting
// @Param module path string true "模块，web database log ..."
// @Accept json
// @Produce json
// @Success 200 {object} ReplyData[map[string]any] 返回配置
// @Router /setting/:module [get]
func settingGet(ctx *gin.Context) {
	api.OK(ctx, viper.GetStringMap(ctx.Param("module")))
}

// @Summary 修改配置
// @Schemes
// @Description 修改配置
// @Tags setting
// @Param module path string true "模块，web database log ..."
// @Param cfg body map[string]any true "配置内容"
// @Accept json
// @Produce json
// @Success 200 {object} ReplyData[int]
// @Router /setting/:module [post]
func settingSet(ctx *gin.Context) {
	m := ctx.Param("module")

	var conf map[string]any
	err := ctx.ShouldBindJSON(&conf)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	for k, v := range conf {
		viper.Set(m+"."+k, v)
	}

	err = config.Store()
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 查询配置表单
// @Schemes
// @Description 查询配置表单
// @Tags setting
// @Param module path string true "模块，web database log ..."
// @Accept json
// @Produce json
// @Success 200 {object} ReplyData[Form] 返回配置表单
// @Router /setting/:module/form [get]
func settingForm(ctx *gin.Context) {
	m := ctx.Param("module")
	md := config.GetModule(m)
	if md == nil {
		api.Fail(ctx, "模块不存在")
		return
	}
	api.OK(ctx, md)
}

// @Summary 查询所有配置
// @Schemes
// @Description 查询所有配置
// @Tags setting
// @Accept json
// @Produce json
// @Success 200 {object} ReplyData[[]Form] 返回配置表单
// @Router /setting/modules [get]
func settingModules(ctx *gin.Context) {
	ms := config.GetModules()
	api.OK(ctx, ms)
}
