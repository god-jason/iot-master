package pool

import (
	"github.com/god-jason/iot-master/pkg/config"
)

// MODULE viper 模块名
const MODULE = "pool"

func init() {
	config.SetDefault(MODULE, "size", 10000)
}

// Info 协程池运行时信息
type Info struct {
	Cap     int `json:"cap,omitempty"`
	Free    int `json:"free,omitempty"`
	Running int `json:"running,omitempty"`
	Waiting int `json:"waiting,omitempty"`
}

// FromViper 从 viper 读 pool 配置
func FromViper() int {
	size := config.GetInt(MODULE, "size")
	if size <= 0 {
		size = 10000
	}
	return size
}
