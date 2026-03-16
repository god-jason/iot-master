package main

import (
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/god-jason/iot-master"
	_ "github.com/god-jason/iot-master/apis"
	"github.com/god-jason/iot-master/iot"
	"github.com/god-jason/iot-master/pkg/boot"
	_ "github.com/god-jason/iot-master/pkg/broker"
	"github.com/god-jason/iot-master/pkg/log"
	"github.com/god-jason/iot-master/pkg/menu"
	_ "github.com/god-jason/iot-master/pkg/oem"
	"github.com/god-jason/iot-master/pkg/page"
	"github.com/god-jason/iot-master/pkg/store"
	_ "github.com/god-jason/iot-master/pkg/table"
	_ "github.com/god-jason/iot-master/pkg/version"
	"github.com/god-jason/iot-master/pkg/web"
	_ "github.com/god-jason/iot-master/weixin"
	"github.com/spf13/viper"
)

func init() {
	//菜单
	menu.File("menu.json")

	//页面
	page.PagesFS = store.Dir("pages")

	//加载协议
	iot.Protocols = store.Dir("protocols")
}

func main() {
	viper.SetConfigName("iot-master")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		//关闭web，出发
		_ = web.Shutdown()
	}()

	//安全退出
	defer boot.Shutdown()

	err := boot.Startup()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = web.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
