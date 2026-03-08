package api

import (
	"github.com/busy-cloud/boat/boot"
)

func init() {
	boot.Register("api", &boot.Task{
		Startup:  Startup, //启动
		Shutdown: nil,
		Depends:  []string{"web", "log", "database"},
	})
}

func Startup() error {

	//if app.Name == "" || app.Name == "boat" {
	//	registerRoutes("api")
	//} else {
	//	registerRoutes("api/" + app.Name) //子目录
	//}

	registerRoutes("api")

	return nil
}
