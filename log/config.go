package log

import (
	"github.com/zgwit/iot-master/config"
)

const MODULE = "log"

func init() {
	config.Register(MODULE, "level", "trace")
	config.Register(MODULE, "caller", true)
	config.Register(MODULE, "text", true)
	config.Register(MODULE, "output", "stdout") //stdout file
	config.Register(MODULE, "filename", "log.txt")
	config.Register(MODULE, "max_size", 10)   //MB
	config.Register(MODULE, "max_backups", 3) //保留文件数
	config.Register(MODULE, "max_age", 30)    //天
	config.Register(MODULE, "compress", true) //gzip压缩

}
