package device

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "iot/device/list", curd.ApiList[Device]())
	api.Register("POST", "iot/device/search", curd.ApiSearch[Device]())
	api.Register("POST", "iot/device/create", curd.ApiCreate[Device]())
	api.Register("GET", "iot/device/:id", curd.ApiGet[Device]())
	api.Register("POST", "iot/device/:id", curd.ApiUpdate[Device]("id", "name", "description", "product_id", "linker_id", "incoming_id", "disabled", "station"))
	api.Register("GET", "iot/device/:id/delete", curd.ApiDelete[Device]())
	api.Register("GET", "iot/device/:id/enable", curd.ApiDisable[Device](false))
	api.Register("GET", "iot/device/:id/disable", curd.ApiDisable[Device](true))

	//物模型
	api.Register("GET", "iot/device/:id/model", curd.ApiGet[DeviceModel]())
	api.Register("POST", "iot/device/:id/model", deviceModelUpdate)

}

func deviceModelUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	var model DeviceModel
	err := ctx.ShouldBind(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	model.Id = id

	_, err = db.Engine().ID(id).Delete(new(DeviceModel)) //不管有没有都删掉
	_, err = db.Engine().ID(id).Insert(&model)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, &model)
}
