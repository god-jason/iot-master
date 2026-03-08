package cron

import (
	"github.com/busy-cloud/boat/boot"
)

func init() {
	boot.Register("cron", &boot.Task{
		Startup:  Startup, //启动
		Shutdown: Shutdown,
		Depends:  []string{},
	})

}

func Startup() (err error) {
	scheduler.Start()
	return
}

func Shutdown() error {
	return scheduler.Shutdown()
}
