package space

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
)

func init() {

	api.Register("POST", "iot/space/count", curd.ApiCount[Space]())

	api.Register("POST", "iot/space/search", curd.ApiSearchWith[Space]([]*curd.With{
		{"space", "parent_id", "id", "name", "parent"},
	}, "id", "name", "project_id", "parent_id", "description", "disabled", "created"))
	api.Register("GET", "iot/space/list", curd.ApiList[Space]())
	api.Register("POST", "iot/space/create", curd.ApiCreate[Space]())
	api.Register("GET", "iot/space/:id", curd.ApiGet[Space]())
	api.Register("POST", "iot/space/:id", curd.ApiUpdate[Space]())
	api.Register("GET", "iot/space/:id/delete", curd.ApiDelete[Space]())
	api.Register("GET", "iot/space/:id/disable", curd.ApiDisable[Space](true))
	api.Register("GET", "iot/space/:id/enable", curd.ApiDisable[Space](false))
}

// @Summary 查询空间
// @Schemes
// @Description 查询空间
// @Tags space
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Space] 返回空间信息
// @Router iot/space/search [post]
func noopSpaceSearch() {}

// @Summary 查询空间
// @Schemes
// @Description 查询空间
// @Tags space
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Space] 返回空间信息
// @Router iot/space/list [get]
func noopSpaceList() {}

// @Summary 创建空间
// @Schemes
// @Description 创建空间
// @Tags space
// @Param search body Space true "空间信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Space] 返回空间信息
// @Router iot/space/create [post]
func noopSpaceCreate() {}

// @Summary 修改空间
// @Schemes
// @Description 修改空间
// @Tags space
// @Param id path int true "空间ID"
// @Param space body Space true "空间信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Space] 返回空间信息
// @Router iot/space/{id} [post]
func noopSpaceUpdate() {}

// @Summary 删除空间
// @Schemes
// @Description 删除空间
// @Tags space
// @Param id path int true "空间ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Space] 返回空间信息
// @Router iot/space/{id}/delete [get]
func noopSpaceDelete() {}

// @Summary 启用空间
// @Schemes
// @Description 启用空间
// @Tags space
// @Param id path int true "空间ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Space] 返回空间信息
// @Router iot/space/{id}/enable [get]
func noopSpaceEnable() {}

// @Summary 禁用空间
// @Schemes
// @Description 禁用空间
// @Tags space
// @Param id path int true "空间ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Space] 返回空间信息
// @Router iot/space/{id}/disable [get]
func noopSpaceDisable() {}
