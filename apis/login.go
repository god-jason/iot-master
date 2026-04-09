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

	//_, _ = db.Engine().InsertOne(&types.UserEvent{UserId: user.id, ModEvent: types.ModEvent{Type: "登录"}})

	//非管理员，都是租户
	if !user.Admin && user.TenantId == "" {
		user.TenantId = user.Id
	}

	//存入session
	session.Set("user", user.Id)
	session.Set("admin", user.Admin)
	session.Set("tenant", user.TenantId)
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
