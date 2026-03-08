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
		if u.Username == "admin" {
			user.Id = "admin"
			user.Name = "管理员"
			user.Admin = true
			_, _ = db.Engine().InsertOne(&user)
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

	//初始化密码
	if !has {
		dp := "123456"
		obj.Password = md5hash(dp)
	}

	if obj.Password != u.Password {
		api.Fail(ctx, "密码错误")
		return
	}

	//生成Token
	token, err := web.JwtGenerate(user.Id, user.Admin, "")
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, gin.H{
		"token": token,
		"user":  &user,
	})
}
