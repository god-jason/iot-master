package log

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/smart"
)

func init() {
	config.Register(MODULE, &config.Form{
		Module: MODULE,
		Title:  "日志配置",
		Fields: []smart.Field{
			{Key: "caller", Label: "显示函数调用", Type: "switch"},
			{Key: "text", Label: "使用文本格式", Type: "switch"},
			{
				Key: "level", Label: "等级", Type: "select", Default: "info",
				Options: []smart.SelectOption{
					{Label: "跟踪 trace", Value: "trace"},
					{Label: "调试 debug", Value: "debug"},
					{Label: "信息 info", Value: "info"},
					{Label: "警告 warn", Value: "warn"},
					{Label: "错误 error", Value: "error"},
					{Label: "严重 fatal", Value: "fatal"},
				},
			},
			{
				Key: "Type", Label: "输出方式", Type: "select", Default: "stdout",
				Options: []smart.SelectOption{
					{Label: "文件", Value: "file"},
					{Label: "多文件", Value: "files"},
					{Label: "标准输出", Value: "stdout"},
				},
			},
			{Key: "filename", Label: "日志文件", Type: "text", Default: "log.txt"},
			{Key: "compress", Label: "日志文件压缩", Type: "switch"},
			{Key: "max_size", Label: "最大尺寸 MB", Type: "number", Default: 10, Min: 1},
			{Key: "max_backups", Label: "保留数量（滚动删除）", Type: "number", Default: 100, Min: 1},
			{Key: "max_age", Label: "最大保留天数", Type: "number", Default: 30, Min: 1},
		},
	})
}
