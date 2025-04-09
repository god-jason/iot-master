package main

import (
	"embed"
	_ "github.com/busy-cloud/boat"
	_ "github.com/busy-cloud/boat-ui"
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/menu"
	"github.com/busy-cloud/boat/page"
	"github.com/busy-cloud/boat/service"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/connector"
	_ "github.com/busy-cloud/influxdb"
	_ "github.com/busy-cloud/modbus"
	_ "github.com/busy-cloud/user"
	"github.com/god-jason/iot-master/protocol"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed pages
var pages embed.FS

//go:embed menus
var menus embed.FS

//go:embed protocols
var protocols embed.FS

func init() {
	//注册页面
	page.EmbedFS(pages, "pages")

	//注册菜单
	menu.EmbedFS(menus, "menus")

	//注册协议
	protocol.EmbedFS(protocols, "protocols")
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
