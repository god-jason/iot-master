package db

import (
	"github.com/zgwit/iot-master/config"
	"xorm.io/xorm"
	"xorm.io/xorm/log"

	//按需加载数据库驱动
	_ "github.com/denisenkom/go-mssqldb" //Sql Server
	_ "github.com/go-sql-driver/mysql"
	//_ "github.com/godror/godror" //Oracle
	_ "github.com/lib/pq" //PostgreSQL
	//_ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3" //CGO版本
	//_ "github.com/glebarez/go-sqlite" //纯Go版本 使用ccgo翻译的，偶有文件锁问题
)

var Engine *xorm.Engine

func Startup() error {
	var err error
	Engine, err = xorm.NewEngine(config.GetString(MODULE, "type"), config.GetString(MODULE, "url"))
	if err != nil {
		return err
	}

	if config.GetBool(MODULE, "debug") {
		Engine.ShowSQL(true)
		Engine.SetLogLevel(log.LOG_DEBUG)
	}
	//Engine.SetLogger(logrus.StandardLogger())

	if config.GetBool(MODULE, "sync") {
		err = Engine.Sync2(models...)
		if err != nil {
			return err
		}
	}

	return nil
}

func Shutdown() error {
	return Engine.Close()
}
