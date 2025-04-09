package app

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

func Proxy(ctx *gin.Context) {
	//app.zipReader = reader

	if ctx.Request.Method != "GET" {
		return
	}

	//插件 前端页面
	if str, has := strings.CutPrefix(ctx.Request.RequestURI, "/app/"); has {
		if len(str) == 0 {
			return
		}

		var app string
		var path string

		if strings.Index(str, "/") > 0 {
			app, path, _ = strings.Cut(str, "/")
		} else {
			app = str
			path = "index.html"
		}

		if p := apps.Load(app); p != nil {
			err := p.ServeFile(path, ctx)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					ctx.String(http.StatusNotFound, "app source not found")
					ctx.Abort()
					return
				}
				_ = ctx.Error(err)
			}
			ctx.Abort()
		}
	}
}
