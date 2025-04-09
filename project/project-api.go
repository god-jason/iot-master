package project

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
)

func init() {

	api.Register("POST", "iot/project/count", curd.ApiCount[Project]())
	api.Register("POST", "iot/project/search", curd.ApiSearch[Project]())
	api.Register("GET", "iot/project/list", curd.ApiList[Project]())
	api.Register("POST", "iot/project/create", curd.ApiCreate[Project]())
	api.Register("GET", "iot/project/:id", curd.ApiGet[Project]())
	api.Register("POST", "iot/project/:id", curd.ApiUpdate[Project]())
	api.Register("GET", "iot/project/:id/delete", curd.ApiDelete[Project]())
	api.Register("GET", "iot/project/:id/disable", curd.ApiDisable[Project](true))
	api.Register("GET", "iot/project/:id/enable", curd.ApiDisable[Project](false))
}

// @Summary 查询项目
// @Schemes
// @Description 查询项目
// @Tags project
// @Param search body curd.ParamSearch true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Project] 返回项目信息
// @Router iot/project/search [post]
func noopProjectSearch() {}

// @Summary 查询项目
// @Schemes
// @Description 查询项目
// @Tags project
// @Param search query curd.ParamList true "查询参数"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyList[Project] 返回项目信息
// @Router iot/project/list [get]
func noopProjectList() {}

// @Summary 创建项目
// @Schemes
// @Description 创建项目
// @Tags project
// @Param search body Project true "项目信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Project] 返回项目信息
// @Router iot/project/create [post]
func noopProjectCreate() {}

// @Summary 修改项目
// @Schemes
// @Description 修改项目
// @Tags project
// @Param id path int true "项目ID"
// @Param project body Project true "项目信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Project] 返回项目信息
// @Router iot/project/{id} [post]
func noopProjectUpdate() {}

// @Summary 删除项目
// @Schemes
// @Description 删除项目
// @Tags project
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Project] 返回项目信息
// @Router iot/project/{id}/delete [get]
func noopProjectDelete() {}

// @Summary 启用项目
// @Schemes
// @Description 启用项目
// @Tags project
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Project] 返回项目信息
// @Router iot/project/{id}/enable [get]
func noopProjectEnable() {}

// @Summary 禁用项目
// @Schemes
// @Description 禁用项目
// @Tags project
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[Project] 返回项目信息
// @Router iot/project/{id}/disable [get]
func noopProjectDisable() {}
