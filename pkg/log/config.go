package log

import (
	"github.com/busy-cloud/boat/config"
)

const MODULE = "log"

func init() {
	config.SetDefault(MODULE, "level", "trace")
	config.SetDefault(MODULE, "caller", true)
	config.SetDefault(MODULE, "text", true)
	config.SetDefault(MODULE, "output", "stdout") //stdout file
	config.SetDefault(MODULE, "filename", "log.txt")
	config.SetDefault(MODULE, "max_size", 10)   //MB
	config.SetDefault(MODULE, "max_backups", 3) //保留文件数
	config.SetDefault(MODULE, "max_age", 30)    //天
	config.SetDefault(MODULE, "compress", true) //gzip压缩

}
