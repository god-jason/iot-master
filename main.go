package main

import (
	"embed"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/god-jason/iot-master/apis"
	"github.com/god-jason/iot-master/iot"
	"github.com/god-jason/iot-master/pkg/boot"
	_ "github.com/god-jason/iot-master/pkg/broker"
	"github.com/god-jason/iot-master/pkg/log"
	"github.com/god-jason/iot-master/pkg/menu"
	_ "github.com/god-jason/iot-master/pkg/oem"
	"github.com/god-jason/iot-master/pkg/page"
	"github.com/god-jason/iot-master/pkg/service"
	"github.com/god-jason/iot-master/pkg/store"
	_ "github.com/god-jason/iot-master/pkg/table"
	"github.com/god-jason/iot-master/pkg/version"
	"github.com/god-jason/iot-master/pkg/web"
	_ "github.com/god-jason/iot-master/weixin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//go:embed dist/browser
var www embed.FS

func init() {
	//菜单
	menu.File("menu.json")

	//页面
	page.PagesFS = store.Dir("pages")

	//加载协议
	iot.Protocols = store.Dir("protocols")

	//前端页面
	web.StaticFS(www, "/", "dist/browser", "index.html")
}

func Startup() error {
	viper.SetConfigName("iot-master")

	err := boot.Startup()
	if err != nil {
		//_ = boot.Shutdown()
		return err
	}

	//异步执行，避免堵塞
	go func() {
		//启动服务
		err := web.Serve()
		if err != nil {
			//安全退出
			//_ = boot.Shutdown()
			log.Error(err)
		}
	}()

	log.Info("main started")

	return nil
}

func Shutdown() error {
	log.Info("main shutdown")

	return boot.Shutdown()
}

func main() {

	help := pflag.BoolP("help", "h", false, "show help")
	install := pflag.BoolP("install", "i", false, "install as service")
	uninstall := pflag.BoolP("uninstall", "u", false, "uninstall service")
	showVersion := pflag.BoolP("version", "v", false, "show version")

	pflag.Parse()
	if *help {
		pflag.PrintDefaults()
		return
	}

	err := service.Register(Startup, Shutdown)
	if err != nil {
		log.Fatal(err)
	}

	if *install {
		log.Info("install service")
		err = service.Install()
		if err != nil {
			log.Fatal(err)
		}
	} else if *uninstall {
		log.Info("uninstall service")
		err = service.Uninstall()
		if err != nil {
			log.Fatal(err)
		}
		return
	} else if *showVersion {
		version.Print()
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Info("signal received ", s)

		//_ = boot.Shutdown()
		err := service.Stop()
		if err != nil {
			log.Error(err)
		}

		time.AfterFunc(10*time.Second, func() {
			os.Exit(0)
		})
	}()

	err = service.Run()
	if err != nil {
		log.Fatal(err)
	}

	println("bye")
}
