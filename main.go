package main

import (
	"github.com/god-jason/bucket/boot"
	"github.com/god-jason/bucket/log"
	"github.com/god-jason/bucket/web"
	"github.com/god-jason/iot-master/api"
)

func main() {
	defer boot.Shutdown()

	err := boot.Startup()
	if err != nil {
		log.Error(err)
		return
	}

	//注册前端接口
	api.RegisterRoutes(web.Engine.Group("/api"))

	//附件
	//web.Engine.Static("/static", "static")
	web.Static.PutDir("/", "www", "/", "index.html")
	//web.Engine.Static("/attach", filepath.Join(viper.GetString("data"), "attach"))

	err = web.Serve()
	if err != nil {
		log.Error(err)
	}
}
