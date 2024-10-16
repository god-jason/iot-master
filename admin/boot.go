package admin

import (
	"github.com/god-jason/bucket/boot"
	"github.com/god-jason/bucket/web"
)

func init() {
	boot.Register("admin", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"config", "web"},
	})
}

func Startup() error {

	web.Engine.POST("api/login", login)

	return nil
}
