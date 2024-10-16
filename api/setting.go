package api

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zgwit/iot-master/config"
	"github.com/zgwit/iot-master/curd"
	"github.com/zgwit/iot-master/setting"
)

// @Summary 查询配置
// @Schemes
// @Description 查询配置
// @Tags setting
// @Param module path string true "模块，web database log ..."
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[map[string]any] 返回配置
// @Router /setting/:module [get]
func settingGet(ctx *gin.Context) {
	curd.OK(ctx, viper.GetStringMap(ctx.Param("module")))
}

// @Summary 修改配置
// @Schemes
// @Description 修改配置
// @Tags setting
// @Param module path string true "模块，web database log ..."
// @Param cfg body map[string]any true "配置内容"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router /setting/:module [post]
func settingSet(ctx *gin.Context) {
	m := ctx.Param("module")

	var conf map[string]any
	err := ctx.ShouldBindJSON(&conf)
	if err != nil {
		curd.Error(ctx, err)
		return
	}
	for k, v := range conf {
		viper.Set(m+"."+k, v)
	}

	err = config.Store()
	if err != nil {
		curd.Error(ctx, err)
		return
	}
	curd.OK(ctx, nil)
}

// @Summary 查询配置表单
// @Schemes
// @Description 查询配置表单
// @Tags setting
// @Param module path string true "模块，web database log ..."
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Module] 返回配置表单
// @Router /setting/:module/form [get]
func settingForm(ctx *gin.Context) {
	m := ctx.Param("module")
	md := setting.Load(m)
	if md == nil {
		curd.Fail(ctx, "模块不存在")
		return
	}
	curd.OK(ctx, md.Form)
}

// @Summary 查询所有配置
// @Schemes
// @Description 查询所有配置
// @Tags setting
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]Module] 返回配置表单
// @Router /setting/modules [get]
func settingModules(ctx *gin.Context) {
	ms := setting.Modules()
	curd.OK(ctx, ms)
}

func init() {
	Register("POST", "setting/:module", settingSet)
	Register("GET", "setting/:module", settingGet)
	Register("GET", "setting/:module/form", settingForm)
	Register("GET", "setting/modules", settingModules)
}
