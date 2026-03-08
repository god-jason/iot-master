package mqtt

import (
	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/lib"
)

const MODULE = "mqtt"

func init() {
	url := "mqtt://localhost:1883"
	//if runtime.GOOS != "windows" {
	//	//使用UnixSocket速度更快
	//	url = "unix://" + os.TempDir() + "/boat.sock" //windows下会出问题，Win10以上虽然支持，但是不能使用绝对路径，因为盘符会被错误解析
	//}
	config.SetDefault(MODULE, "url", url)
	config.SetDefault(MODULE, "clientId", lib.RandomString(12))
	config.SetDefault(MODULE, "username", "")
	config.SetDefault(MODULE, "password", "")
}
