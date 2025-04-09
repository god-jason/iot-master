package project

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"xorm.io/xorm/schemas"
)

func init() {
	api.Register("GET", "iot/project/:id/app/list", projectAppList)
	api.Register("GET", "iot/project/:id/app/:app/exists", projectAppExists)
	api.Register("GET", "iot/project/:id/app/:app/bind", projectAppBind)
	api.Register("GET", "iot/project/:id/app/:app/unbind", projectAppUnbind)
	api.Register("GET", "iot/project/:id/app/:app/disable", projectAppDisable)
	api.Register("GET", "iot/project/:id/app/:app/enable", projectAppEnable)
}

// @Summary 项目应用列表
// @Schemes
// @Description 项目应用列表
// @Tags project-app
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]ProjectApp] 返回项目应用信息
// @Router iot/project/{id}/app/list [get]
func projectAppList(ctx *gin.Context) {
	var pds []ProjectApp
	err := db.Engine().Where("project_id=?", ctx.Param("id")).Find(&pds)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, pds)
}

// @Summary 判断项目应用是否存在
// @Schemes
// @Description 判断项目应用是否存在
// @Tags project-app
// @Param id path int true "项目ID"
// @Param app path int true "应用ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[bool]
// @Router iot/project/{id}/app/{app}/exists [get]
func projectAppExists(ctx *gin.Context) {
	pd := ProjectApp{
		ProjectId: ctx.Param("id"),
		AppId:     ctx.Param("app"),
	}
	has, err := db.Engine().Exist(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, has)
}

// @Summary 绑定项目应用
// @Schemes
// @Description 绑定项目应用
// @Tags project-app
// @Param id path int true "项目ID"
// @Param app path int true "应用ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/app/{app}/bind [get]
func projectAppBind(ctx *gin.Context) {
	pd := ProjectApp{
		ProjectId: ctx.Param("id"),
		AppId:     ctx.Param("app"),
	}
	_, err := db.Engine().InsertOne(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 删除项目应用
// @Schemes
// @Description 删除项目应用
// @Tags project-app
// @Param id path int true "项目ID"
// @Param app path int true "应用ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/app/{app}/unbind [get]
func projectAppUnbind(ctx *gin.Context) {
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("app")}).Delete(new(ProjectApp))
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 禁用项目应用
// @Schemes
// @Description 禁用项目应用
// @Tags project-app
// @Param id path int true "项目ID"
// @Param app path int true "应用ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/app/{app}/disable [get]
func projectAppDisable(ctx *gin.Context) {
	pd := ProjectApp{Disabled: true}
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("app")}).Cols("disabled").Update(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 启用项目应用
// @Schemes
// @Description 启用项目应用
// @Tags project-app
// @Param id path int true "项目ID"
// @Param app path int true "应用ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/app/{app}/enable [get]
func projectAppEnable(ctx *gin.Context) {
	pd := ProjectApp{Disabled: false}
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("app")}).Cols("disabled").Update(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}
