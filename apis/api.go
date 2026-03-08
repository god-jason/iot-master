package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
)

func init() {

	api.RegisterUnAuthorized("POST", "login", login)
	//api.RegisterUnAuthorized("POST", "login", login)

	api.RegisterUnAuthorized("POST", "auth", auth)
	//api.RegisterUnAuthorized("GET", "auth", auth)

	api.Register("GET", "logout", logout)
	//api.Register("GET", "logout", logout)

	api.Register("POST", "password", password)
	//api.Register("POST", "password", password)

	api.Register("GET", "me", userMe)
	//api.Register("GET", "me", userMe)

	api.Register("POST", "password/:id", userPassword)
}

func userMe(ctx *gin.Context) {
	id := ctx.GetString("user")

	var user User
	has, err := db.Engine().ID(id).Get(&user)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	if !has {
		api.Fail(ctx, "用户不存在")
		return
	}

	api.OK(ctx, &user)
}

type userPasswordObj struct {
	Password string `json:"password"`
}

func userPassword(ctx *gin.Context) {

	var obj userPasswordObj
	if err := ctx.ShouldBindJSON(&obj); err != nil {
		api.Error(ctx, err)
		return
	}

	var p Password
	p.Id = ctx.Param("id")
	p.Password = md5hash(obj.Password)

	_, _ = db.Engine().ID(p.Id).Delete(new(Password)) //不管有没有都删掉
	_, err := db.Engine().InsertOne(&p)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
