package weixin

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/web"
	"github.com/rs/xid"
)

func init() {
	api.RegisterUnAuthorized("GET", "weixin/code2session", code2session)
}

func code2session(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		api.Fail(ctx, "缺少code")
		return
	}

	resp, err := mp.Auth.Session(context.Background(), code)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	if resp.OpenID == "" {
		api.Fail(ctx, "获取OpenId失败")
		return
	}

	var user User
	has, err := db.Engine().Where("openid=?", resp.OpenID).Get(&user)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	if !has {
		//user.Id = "wx_" + resp.OpenID
		user.Id = xid.New().String()
		user.Name = "微信用户"
		user.OpenId = resp.OpenID
		user.UnionId = resp.UnionID
		_, err = db.Engine().Insert(&user)
		if err != nil {
			api.Error(ctx, err)
			return
		}
	} else {
		if user.Disabled {
			api.Fail(ctx, "用户被禁用")
			return
		}
	}

	token, err := web.JwtGenerate(user.Id, user.Admin, "")
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, gin.H{"token": token, "user": user, "session": resp.SessionKey})
}
