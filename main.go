package main

import (
	"github.com/zgwit/iot-master/api"
	"github.com/zgwit/iot-master/boot"
	"github.com/zgwit/iot-master/log"
	"github.com/zgwit/iot-master/web"
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
