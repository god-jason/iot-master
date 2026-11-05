package broker

import (
	"fmt"

	beeq "github.com/god-jason/iot-master/pkg/beeq"
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/logger"
)

var (
	server  *beeq.Server // 内嵌 broker 服务端实例
	service *Service     // 消息服务，Device 出站指令复用同一实例
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}

// Startup 启动内嵌 MQTT broker（enable=false 时跳过）
func Startup() error {
	if !Enabled() {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", config.GetString(MODULE, "host"), config.GetInt(MODULE, "port"))

	service = NewService()
	server = beeq.NewServer(beeq.DefaultOptions(), service)
	service.Attach(server)

	// 后台监听，避免阻塞 boot 启动流程
	go func() {
		logger.Info("内嵌 MQTT Broker 启动中", "address", addr)
		if err := server.Listen("tcp://" + addr); err != nil {
			logger.Error("内嵌 MQTT Broker 启动失败", "error", err)
		}
	}()

	return nil
}

// Shutdown 关闭内嵌 MQTT broker
func Shutdown() error {
	if server != nil {
		server.Close()
		server = nil
	}
	service = nil
	return nil
}

// GetService 获取内嵌 broker 消息服务实例（未启用返回 nil）
func GetService() *Service { return service }
