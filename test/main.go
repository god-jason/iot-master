package main

import (
	_ "github.com/busy-cloud/boat/apis"
	"github.com/busy-cloud/boat/apps"
	"github.com/busy-cloud/boat/boot"
	_ "github.com/busy-cloud/boat/broker"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/boat/table"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/connector"
	_ "github.com/busy-cloud/dash"
	_ "github.com/busy-cloud/influxdb"
	_ "github.com/busy-cloud/modbus"
	_ "github.com/busy-cloud/noob"
	_ "github.com/busy-cloud/user"
	_ "github.com/god-jason/iot-master"
	"github.com/god-jason/iot-master/protocol"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	//测试
	apps.Pages().Dir("pages", "")

	//协议
	protocol.Dir("protocols", "")

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
