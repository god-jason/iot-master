package apps

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Proxy(ctx *gin.Context) {

	if str, has := strings.CutPrefix(ctx.Request.RequestURI, "/app/"); has {
		if len(str) == 0 {
			return
		}
		if app, _, has := strings.Cut(str, "/"); has {
			if p := _apps.Load(app); p != nil {
				if p.ServeApi(ctx) {
					ctx.Abort()
					return
				}
			}
		}
	}

	ctx.Next()
}
