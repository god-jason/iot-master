package gateway

import (
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
)

// MODULE 启动模块名
const MODULE = "gateway"

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config", "db"},
	})
}

// Startup 启动所有配置为 enabled 的 Gateway Server
func Startup() error {
	// 软网关依赖 DB 读取配置，数据库未就绪时跳过（由 boot 依赖顺序保证通常不会命中）
	if db.Engine == nil {
		logger.Warn("软网关跳过启动：数据库未初始化")
		return nil
	}

	var servers []types.GatewayServer
	if err := db.Engine.Find(&servers); err != nil {
		logger.Warn("软网关读取 gateway_server 失败", "error", err)
		return nil
	}

	// 逐个启动已启用配置；单个失败不影响其它 Server
	for i := range servers {
		cfg := &servers[i]
		if cfg.Disabled {
			continue
		}
		if _, err := GetGatewayManager().StartServer(cfg); err != nil {
			logger.Error("软网关 Server 启动失败", "server_id", cfg.ID, "name", cfg.Name, "error", err)
		}
	}
	return nil
}

// Shutdown 关闭所有 Gateway Server
func Shutdown() error {
	GetGatewayManager().ShutdownAll()
	return nil
}
