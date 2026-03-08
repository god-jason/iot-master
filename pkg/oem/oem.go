package oem

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/config"
	"github.com/gin-gonic/gin"
)

func init() {
	api.RegisterUnAuthorized("GET", "oem", oem)
	api.RegisterAdmin("POST", "oem/update", oemUpdate)
}

type OEM struct {
	Name string `json:"name,omitempty"`
	Logo string `json:"logo,omitempty"`
}

// @Summary 获取oem信息
// @Schemes
// @Description 获取oem信息
// @Tags oem
// @Accept json
// @Produce json
// @Success 200 {object} ReplyData[OEM] 返回信息
// @Router /oem [get]
func oem(ctx *gin.Context) {
	api.OK(ctx, OEM{
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
// @Success 200 {object} ReplyData[int] 返回nil
// @Router /oem [post]
func oemUpdate(ctx *gin.Context) {
	var oem OEM
	err := ctx.ShouldBindJSON(&oem)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}
