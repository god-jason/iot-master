package web

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/exception"
	"github.com/busy-cloud/boat/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var engine *gin.Engine
var Server *http.Server

func Engine() *gin.Engine {
	return engine
}

func Startup() error {
	if !config.GetBool(MODULE, "debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	//GIN初始化
	//engine := gin.Default()
	engine = gin.New()
	engine.Use(gin.Recovery())

	if config.GetBool(MODULE, "debug") {
		engine.Use(gin.Logger())
	}

	//跨域问题
	if config.GetBool(MODULE, "cors") {
		c := cors.DefaultConfig()
		c.AllowAllOrigins = true
		c.AllowCredentials = true
		engine.Use(cors.New(c))
	}

	//启用session
	engine.Use(sessions.Sessions("boat", cookie.NewStore([]byte("boat"))))

	//开启压缩
	if config.GetBool(MODULE, "gzip") {
		engine.Use(gzip.Gzip(gzip.DefaultCompression)) //gzip.WithExcludedPathsRegexs([]string{".*"})
	}

	JwtKey = []byte(config.GetString(MODULE, "jwt_key"))
	JwtExpire = time.Hour * time.Duration(config.GetInt(MODULE, "jwt_expire"))

	//engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return nil
}

func Shutdown() error {
	if Server == nil {
		return exception.New("服务未启动")
	}
	return Server.Shutdown(context.Background())
}

func Serve() error {

	//静态文件
	tm := time.Now()
	engine.Use(func(c *gin.Context) {

		//如果已经写过内容，就不用再写了
		if c.Writer.Written() {
			return
		}

		if c.Request.Method == http.MethodGet {
			f, err := OpenStaticFile(c.Request.URL.Path)
			if err != nil {
				c.Next()
				return
			}

			defer f.Close()
			stat, err := f.Stat()
			if err != nil {
				//c.Next() //500错误
				c.Error(err)
				return
			}

			if !stat.IsDir() {
				fn := c.Request.URL.Path
				//fn := c.Request.URL.Path + ".html" //避免DetectContentType
				http.ServeContent(c.Writer, c.Request, fn, tm, f)
				return
			}

		}

		c.Next()
	})

	//按不同模式启动
	mode := config.GetString(MODULE, "mode")
	switch strings.ToLower(mode) {
	case "http", "tcp":
		return ServeHTTP()
	case "https", "tls", "ssl":
		return ServeTLS()
	case "autocert", "letsencrypt":
		return ServeAutoCert()
	case "unix", "socket":
		return ServeUnix()
	default:
		return ServeHTTP()
	}
}

func ServeHTTP() error {
	port := config.GetInt(MODULE, "port")
	addr := ":" + strconv.Itoa(port)
	log.Info("web ", port)
	//return engine.Run(addr)
	Server = &http.Server{Addr: addr, Handler: engine.Handler()}
	return Server.ListenAndServe()
}

func ServeUnix() error {
	socket := config.GetString(MODULE, "unix_socket")
	log.Info("web ", socket)
	Server = &http.Server{Addr: socket, Handler: engine.Handler()}
	ln, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	return Server.Serve(ln)
}
