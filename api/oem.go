package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/config"
	"github.com/zgwit/iot-master/web/curd"
)

type OEM struct {
	Name string `json:"name"`
	Logo string `json:"logo"`
}

// @Summary 获取oem信息
// @Schemes
// @Description 获取oem信息
// @Tags oem
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[OEM] 返回信息
// @Router /oem [get]
func oem(ctx *gin.Context) {
	curd.OK(ctx, OEM{
		Name: config.GetString("oem", "name"),
		Logo: config.GetString("oem", "logo"),
	})
}

// @Summary 修改oem信息
// @Schemes
// @Description 修改oem信息
// @Tags oem
// @Param search body OEM true "信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int] 返回nil
// @Router /oem [post]
func oemUpdate(ctx *gin.Context) {
	var oem OEM
	err := ctx.ShouldBindJSON(&oem)
	if err != nil {
		curd.Error(ctx, err)
		return
	}
	curd.OK(ctx, nil)
}

func oemRouter(app *gin.RouterGroup) {
	app.POST("", oemUpdate)
}
