package apis

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
)

type loginObj struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

func md5hash(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func login(ctx *gin.Context) {
	session := sessions.Default(ctx)

	var obj loginObj
	if err := ctx.ShouldBindJSON(&obj); err != nil {
		api.Error(ctx, err)
		return
	}

	var user User
	has, err := db.Engine().Where("id=?", obj.Username).Get(&user)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	if !has {
		//管理员自动创建
		if obj.Username == "admin" {
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

	var password Password
	has, err = db.Engine().ID(user.Id).Get(&password)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//初始化密码
	if !has {
		dp := "123456"

		password.Id = user.Id
		password.Password = md5hash(dp)
		_, err = db.Engine().InsertOne(&password)
		if err != nil {
			api.Error(ctx, err)
			return
		}
	}

	if password.Password != obj.Password {
		api.Fail(ctx, "密码错误")
		return
	}

	//_, _ = db.Engine().InsertOne(&types.UserEvent{UserId: user.id, ModEvent: types.ModEvent{Type: "登录"}})

	//存入session
	session.Set("user", user.Id)
	session.Set("admin", user.Admin)
	_ = session.Save()

	api.OK(ctx, user)
}

func logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	u := session.Get("user")
	if u == nil {
		api.Fail(ctx, "未登录")
		return
	}

	//user := u.(int64)
	//_, _ = db.Engine().InsertOne(&types.UserEvent{UserId: user, ModEvent: types.ModEvent{Type: "退出"}})

	session.Clear()
	_ = session.Save()
	api.OK(ctx, nil)
}

type passwordObj struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func password(ctx *gin.Context) {

	var obj passwordObj
	if err := ctx.ShouldBindJSON(&obj); err != nil {
		api.Error(ctx, err)
		return
	}

	var pwd Password
	has, err := db.Engine().ID(ctx.GetString("user")).Get(&pwd)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	if !has {
		pwd.Id = ctx.GetString("user")
		pwd.Password = obj.New //前端已经加密过
		_, err = db.Engine().InsertOne(&pwd)
		if err != nil {
			api.Error(ctx, err)
			return
		}
	} else {
		if obj.Old != pwd.Password {
			api.Fail(ctx, "密码错误")
			return
		}

		pwd.Password = md5hash(obj.New)
		_, err = db.Engine().ID(ctx.GetString("user")).Cols("password").Update(&pwd)
		if err != nil {
			api.Error(ctx, err)
			return
		}
	}

	api.OK(ctx, nil)
}
