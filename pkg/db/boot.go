package db

import (
	"fmt"

	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"config"},
	})
}

var Engine *xorm.Engine

// InitEngine 显式初始化数据库引擎（供 install/setup 等独立场景手动连库测试用）
func InitEngine(driver, dsn string) error {
	engine, err := xorm.NewEngine(driver, dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	engine.SetMapper(names.GonicMapper{})
	Engine = engine
	logger.Info("数据库连接成功", "driver", driver)
	return nil
}

// SyncTables 手动同步表结构（需 Engine 已 Init）
func SyncTables(models ...interface{}) error {
	if Engine == nil {
		return fmt.Errorf("Engine未初始化")
	}
	if err := Engine.Sync2(models...); err != nil {
		return err
	}
	logger.Info("数据表同步完成", "count", len(models))
	return nil
}

// Startup boot 启动入口：从 viper 读 type/url，连库，可选 debug + 可选 sync
func Startup() error {
	if !Enabled() {
		logger.Info("db 模块已禁用，跳过启动")
		return nil
	}

	// 支持两种配置写法：
	//   1. 分散字段（master.yaml）：db.driver/host/port/database/username/password
	//   2. 单连接串（旧写法）：db.type + db.url
	driver := config.GetString(MODULE, "driver")
	var dsn string
	if driver != "" {
		var err error
		dsn, err = DSN(
			driver,
			config.GetString(MODULE, "host"),
			config.GetInt(MODULE, "port"),
			config.GetString(MODULE, "database"),
			config.GetString(MODULE, "username"),
			config.GetString(MODULE, "password"),
			config.GetString(MODULE, "sslmode"),
			config.GetString(MODULE, "params"),
		)
		if err != nil {
			return fmt.Errorf("数据库连接失败: %w", err)
		}
	} else {
		driver = config.GetString(MODULE, "type")
		dsn = config.GetString(MODULE, "url")
	}

	engine, err := xorm.NewEngine(driver, dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}
	engine.SetMapper(names.GonicMapper{})

	if config.GetBool(MODULE, "debug") {
		engine.ShowSQL(true)
		engine.SetLogLevel(log.LOG_DEBUG)
	}

	if config.GetBool(MODULE, "sync") {
		if err := engine.Sync2(models...); err != nil {
			logger.Warn("自动同步表结构失败", "error", err)
		}
	}

	Engine = engine
	logger.Info("数据库连接成功", "driver", driver)
	return nil
}

// Shutdown boot 关闭入口
func Shutdown() error {
	if Engine != nil {
		Engine.Close()
		Engine = nil
	}
	return nil
}

// Close 兼容老调用方
func Close() { _ = Shutdown() }
