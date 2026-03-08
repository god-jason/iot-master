package apis

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
	"github.com/god-jason/iot-master/pkg/version"
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
