package web

import (
	"math/rand/v2"

	"github.com/busy-cloud/boat/config"
)

const MODULE = "web"

func init() {
	config.SetDefault(MODULE, "mode", "http")               //http, https, tls, ssl, autocert, letsencrypt, unix
	config.SetDefault(MODULE, "port", 8000+rand.UintN(100)) //端口号 8000 - 8099
	config.SetDefault(MODULE, "debug", false)
	config.SetDefault(MODULE, "cors", false)
	config.SetDefault(MODULE, "gzip", true)
	config.SetDefault(MODULE, "tls_cert", "")
	config.SetDefault(MODULE, "tls_key", "")
	config.SetDefault(MODULE, "hosts", []string{}) //域名
	config.SetDefault(MODULE, "email", "")
	config.SetDefault(MODULE, "jwt_key", "boat")
	config.SetDefault(MODULE, "jwt_expire", 24*30) //小时
	config.SetDefault(MODULE, "unix_socket", "")

	//可以通过环境变量，强制修改监听类型和端口
	//BOAT_WEB.MODE=unix;
	//BOAT_WEB.UNIX_SOCKET=boat.sock

}
