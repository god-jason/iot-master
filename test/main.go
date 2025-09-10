package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/busy-cloud/boat-ui"
	_ "github.com/busy-cloud/boat/apis"
	"github.com/busy-cloud/boat/apps"
	"github.com/busy-cloud/boat/boot"
	_ "github.com/busy-cloud/boat/broker"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/boat/oem"
	"github.com/busy-cloud/boat/store"
	_ "github.com/busy-cloud/boat/table"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/dash"
	_ "github.com/busy-cloud/influxdb"
	_ "github.com/busy-cloud/modbus"
	_ "github.com/busy-cloud/saas"
	_ "github.com/busy-cloud/serial-port"
	_ "github.com/busy-cloud/tcp-client"
	_ "github.com/busy-cloud/tcp-server"
	_ "github.com/god-jason/iot-master"
	"github.com/spf13/viper"
)

func init() {
	manifest, err := os.ReadFile("manifest.json")
	if err != nil {
		log.Fatal(err)
	}

	//注册为内部插件
	var a apps.App
	err = json.Unmarshal(manifest, &a)
	if err != nil {
		log.Fatal(err)
	}
	apps.Register(&a)

	//注册资源
	a.AssetsFS = store.Dir("assets")
	a.PagesFS = store.Dir("pages")
	a.TablesFS = store.Dir("tables")
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
