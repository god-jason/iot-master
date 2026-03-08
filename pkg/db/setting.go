package db

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/smart"
)

func init() {
	config.Register(MODULE, &config.Form{
		Title:  "数据库配置",
		Module: MODULE,
		Fields: []smart.Field{
			{
				Key: "Type", Label: "数据库类型", Type: "select", Default: "mysql",
				Options: []smart.SelectOption{
					{Label: "SQLite（内置）", Value: "sqlite3"},
					{Label: "MySQL", Value: "mysql"},
					{Label: "Postgres SQL", Value: "postgres"},
					//{Label: "MS SQL Server", Value: "sqlserver"},
					//{Label: "Oracle", Value: "godror"},
				},
			},
			{Key: "url", Label: "连接字符串", Type: "text"},
			{Key: "debug", Label: "调试模式", Type: "switch"},
			{Key: "sync", Label: "自动创建表结构", Type: "switch"},
		},
	})
}
