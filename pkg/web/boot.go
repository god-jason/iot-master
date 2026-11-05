package web

import (
	"github.com/god-jason/iot-master/pkg/boot"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		// Startup 不自动启动 Web 服务——需要外部注入路由后 RunServer，所以这里留空
		Startup:  func() error { return nil },
		Shutdown: func() error { return nil },
		Depends:  []string{"config"},
	})
}
