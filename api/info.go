package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/pkg/build"
	"github.com/zgwit/iot-master/web/curd"
	"runtime"
)

func info(ctx *gin.Context) {
	curd.OK(ctx, gin.H{
		"version": build.Version,
		"build":   build.Build,
		"git":     build.GitHash,
		"gin":     gin.Version,
		"runtime": runtime.Version(),
	})
}
