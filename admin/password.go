package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/bucket/config"
	"github.com/god-jason/iot-master/curd"
)

type passwordObj struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func password(ctx *gin.Context) {

	var obj passwordObj
	if err := ctx.ShouldBind(&obj); err != nil {
		curd.Error(ctx, err)
		return
	}

	if obj.Old != config.GetString(MODULE, "password") {
		curd.Fail(ctx, "密码错误")
		return
	}

	//更新密码
	config.Set(MODULE, "password", obj.New)

	err := config.Store()
	if err != nil {
		curd.Error(ctx, err)
		return
	}

	curd.OK(ctx, nil)
}
