package memstore

import (
	"fmt"

	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}

// Startup boot 启动入口：
//   - enable=false → 直接用内存 store，不尝试连 Redis
//   - enable=true  → 按 viper 读的 host/port/password/db 连 Redis，失败降级内存
func Startup() error {
	addr := ""
	password := ""
	dbNum := 0

	if Enabled() {
		host := config.GetString(MODULE, "host")
		port := config.GetInt(MODULE, "port")
		password = config.GetString(MODULE, "password")
		dbNum = config.GetInt(MODULE, "db")
		addr = fmt.Sprintf("%s:%d", host, port)
	}

	return Init(addr, password, dbNum)
}

// Shutdown boot 关闭入口
func Shutdown() error {
	Close()
	return nil
}
