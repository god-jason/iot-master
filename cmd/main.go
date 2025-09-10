package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/busy-cloud/boat-ui"
	_ "github.com/busy-cloud/boat/apis"
	_ "github.com/busy-cloud/boat/apps"
	"github.com/busy-cloud/boat/boot"
	_ "github.com/busy-cloud/boat/broker"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/boat/oem"
	"github.com/busy-cloud/boat/service"
	_ "github.com/busy-cloud/boat/table"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/dash"
	_ "github.com/busy-cloud/influxdb"
	_ "github.com/busy-cloud/modbus"
	_ "github.com/busy-cloud/tcp-server"
	_ "github.com/busy-cloud/user"
	_ "github.com/god-jason/iot-master" //主程序
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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
