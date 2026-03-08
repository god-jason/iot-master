package apis

import (
	"runtime"

	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/version"
	"github.com/gin-gonic/gin"
)

func init() {
	api.RegisterUnAuthorized("GET", "version", info)
}

func info(ctx *gin.Context) {
	api.OK(ctx, gin.H{
		"runtime": runtime.Version(),
		"build":   version.Build,
		"version": version.Version,
		"git":     version.GitHash,
		"gin":     gin.Version,
	})
}
