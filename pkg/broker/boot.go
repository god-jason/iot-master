package broker

import (
	"log/slog"
	"net"
	"net/url"
	"os"

	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	client "github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/boat/web"
	paho "github.com/eclipse/paho.mqtt.golang"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

func init() {
	boot.Register("broker", &boot.Task{
		Startup:  Startup, //启动
		Shutdown: Shutdown,
		Depends:  []string{"web", "log", "database"},
	})

}

var server *mqtt.Server

func Startup() (err error) {
	//禁用不启动
	if !config.GetBool(MODULE, "enable") {
		return nil
	}

	//解析日志等级
	loglevel := config.GetString(MODULE, "loglevel")
	var level slog.Level
	err = level.UnmarshalText([]byte(loglevel))
	if err != nil {
		level = slog.LevelWarn
	}

	//创建服务
	opts := &mqtt.Options{
		InlineClient: true,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})),
	}
	server = mqtt.New(opts)

	//匿名
	if config.GetBool(MODULE, "anonymous") {
		err = server.AddHook(new(auth.AllowHook), nil)
	} else {
		err = server.AddHook(new(Hook), nil)
	}

	if err != nil {
		return err
	}

	//启用unixsock，速度更快
	if unixsock := config.GetString(MODULE, "unixsock"); unixsock != "" {
		err = os.Remove(unixsock)
		if err != nil {
			if !os.IsNotExist(err) { //TODO 此处没有正常执行
				//文件被占用
				log.Fatal("Boat不能重复启动")
			}
		}

		err = server.AddListener(listeners.NewUnixSock(listeners.Config{
			ID:      "unix",
			Address: unixsock,
		}))
		if err != nil {
			return err
		}
	}

	//内置监听
	err = server.AddListener(listeners.NewTCP(listeners.Config{
		ID:      "base",
		Address: ":" + config.GetString(MODULE, "port"),
	}))
	if err != nil {
		return err
	}

	//监听Websocket
	web.Engine().GET("/mqtt", GinBridge)

	//向mqtt客户端注册内部连接方式
	client.CustomConnectionFunc = func(uri *url.URL, options paho.ClientOptions) (net.Conn, error) {
		c1, c2 := lib.NewVConn()
		//EstablishConnection会读取connect，导致拥堵
		go func() {
			err := server.EstablishConnection("internal", c1)
			if err != nil {
				log.Error(err)
			}
		}()
		return c2, nil
	}

	return server.Serve()
}

func Shutdown() error {
	if server != nil {
		return server.Close()
	}
	return nil
}
