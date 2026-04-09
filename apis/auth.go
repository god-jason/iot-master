package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/web"
)

func auth(ctx *gin.Context) {
	var u loginObj
	if err := ctx.ShouldBindJSON(&u); err != nil {
		api.Error(ctx, err)
		return
	}

	var user User
	has, err := db.Engine().Where("id=?", u.Username).Get(&user)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	if !has {
		//管理员自动创建
		if u.Username == "admin" {
			user.Id = "admin"
			user.Name = "管理员"
			user.Admin = true
			_, _ = db.Engine().InsertOne(&user)

			//初始化管理员密码
			var pas Password
			pas.Id = user.Id
			pas.Password = md5hash("123456") //管理默认密码
			_, _ = db.Engine().InsertOne(&pas)
		} else {
			api.Fail(ctx, "找不到用户")
			return
		}
	}

	if user.Disabled {
		api.Fail(ctx, "用户已禁用")
		return
	}

	var obj Password
	has, err = db.Engine().ID(user.Id).Get(&obj)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	if !has {
		api.Fail(ctx, "用户未初始化密码")
		return
	}

	if obj.Password != u.Password {
		api.Fail(ctx, "用户名或密码错误")
		return
	}

	//非管理员，都是租户
	if !user.Admin && user.TenantId == "" {
		user.TenantId = user.Id
	}

	//生成Token
	token, err := web.JwtGenerate(user.Id, user.Admin, user.TenantId)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, gin.H{
		"token": token,
		"user":  &user,
	})
}
