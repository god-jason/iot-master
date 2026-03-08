package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/god-jason/iot-master"
	_ "github.com/god-jason/iot-master/apis"
	"github.com/god-jason/iot-master/apps"
	"github.com/god-jason/iot-master/iot"
	"github.com/god-jason/iot-master/pkg/boot"
	_ "github.com/god-jason/iot-master/pkg/broker"
	"github.com/god-jason/iot-master/pkg/log"
	_ "github.com/god-jason/iot-master/pkg/oem"
	"github.com/god-jason/iot-master/pkg/store"
	_ "github.com/god-jason/iot-master/pkg/table"
	_ "github.com/god-jason/iot-master/pkg/version"
	"github.com/god-jason/iot-master/pkg/web"
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
