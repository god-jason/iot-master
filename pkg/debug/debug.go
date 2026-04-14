package debug

import (
	"github.com/gin-contrib/pprof"
	"github.com/god-jason/iot-master/pkg/web"
)

func Startup() error {
	//注册到路由上
	pprof.Register(web.Engine())
	return nil
}
