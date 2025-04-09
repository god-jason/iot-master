package project

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/db"
	"github.com/gin-gonic/gin"
	"xorm.io/xorm/schemas"
)

func init() {
	api.Register("GET", "iot/project/:id/user/list", projectUserList)
	api.Register("GET", "iot/project/:id/user/:user/exists", projectUserExists)
	api.Register("GET", "iot/project/:id/user/:user/bind", projectUserBind)
	api.Register("GET", "iot/project/:id/user/:user/unbind", projectUserUnbind)
	api.Register("GET", "iot/project/:id/user/:user/disable", projectUserDisable)
	api.Register("GET", "iot/project/:id/user/:user/enable", projectUserEnable)
	api.Register("POST", "iot/project/:id/user/:user", projectUserUpdate)
	//个人项目
	api.Register("GET", "/user/:id/projects", userProjects)
}

// @Summary 项目用户列表
// @Schemes
// @Description 项目用户列表
// @Tags project-user
// @Param id path int true "项目ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]ProjectUser] 返回项目用户信息
// @Router iot/project/{id}/user/list [get]
func projectUserList(ctx *gin.Context) {
	var pds []ProjectUser
	err := db.Engine().
		Select("project_user.project_id, project_user.user_id, project_user.admin, project_user.disabled, project_user.created, user.name as user").
		Join("INNER", "user", "user.id=project_user.user_id").
		Where("project_user.project_id=?", ctx.Param("id")).
		Find(&pds)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, pds)
}

// @Summary 判断项目用户是否存在
// @Schemes
// @Description 判断项目用户是否存在
// @Tags project-user
// @Param id path int true "项目ID"
// @Param user path int true "用户ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[bool]
// @Router iot/project/{id}/user/{user}/exists [get]
func projectUserExists(ctx *gin.Context) {
	pd := ProjectUser{
		ProjectId: ctx.Param("id"),
		UserId:    ctx.Param("user"),
	}
	has, err := db.Engine().Exist(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, has)
}

// @Summary 绑定项目用户
// @Schemes
// @Description 绑定项目用户
// @Tags project-user
// @Param id path int true "项目ID"
// @Param user path int true "用户ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/user/{user}/bind [get]
func projectUserBind(ctx *gin.Context) {
	pd := ProjectUser{
		ProjectId: ctx.Param("id"),
		UserId:    ctx.Param("user"),
	}
	_, err := db.Engine().InsertOne(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 删除项目用户
// @Schemes
// @Description 删除项目用户
// @Tags project-user
// @Param id path int true "项目ID"
// @Param user path int true "用户ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/user/{user}/unbind [get]
func projectUserUnbind(ctx *gin.Context) {
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("user")}).Delete(new(ProjectUser))
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 禁用项目用户
// @Schemes
// @Description 禁用项目用户
// @Tags project-user
// @Param id path int true "项目ID"
// @Param user path int true "用户ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/user/{user}/disable [get]
func projectUserDisable(ctx *gin.Context) {
	pd := ProjectUser{Disabled: true}
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("user")}).Cols("disabled").Update(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 启用项目用户
// @Schemes
// @Description 启用项目用户
// @Tags project-user
// @Param id path int true "项目ID"
// @Param user path int true "用户ID"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/user/{user}/enable [get]
func projectUserEnable(ctx *gin.Context) {
	pd := ProjectUser{Disabled: false}
	_, err := db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("user")}).Cols("disabled").Update(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 修改项目用户
// @Schemes
// @Description 修改项目用户
// @Tags project-user
// @Param id path int true "项目ID"
// @Param user path int true "用户ID"
// @Param project-user body ProjectUser true "项目用户信息"
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[int]
// @Router iot/project/{id}/user/{user} [post]
func projectUserUpdate(ctx *gin.Context) {
	var pd ProjectUser
	err := ctx.ShouldBindJSON(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	_, err = db.Engine().ID(schemas.PK{ctx.Param("id"), ctx.Param("user")}).
		Cols("user_id", "disabled").
		Update(&pd)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	api.OK(ctx, nil)
}

// @Summary 获取用户的项目列表
// @Schemes
// @Description 获取用户的项目列表
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} curd.ReplyData[[]Project] 返回项目列表
// @Router /user/{id}iot/projects [get]
func userProjects(ctx *gin.Context) {
	id := ctx.GetString("id")

	var projects []*Project
	err := db.Engine().Join("INNER", "project_user", "project_user.project_id=id").
		Where("project_user.user_id=?", id).Find(&projects)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, projects)
}
