package web

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/zgwit/iot-master/config"
	"github.com/zgwit/iot-master/log"
	"github.com/zgwit/iot-master/pkg/exception"
	"net/http"
	"path"
	"strconv"
	"time"
)

var Engine *gin.Engine
var Server *http.Server

func Startup() error {
	if !config.GetBool(MODULE, "debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	//GIN初始化
	//Engine := gin.Default()
	Engine = gin.New()
	Engine.Use(gin.Recovery())

	if config.GetBool(MODULE, "debug") {
		Engine.Use(gin.Logger())
	}

	//跨域问题
	if config.GetBool(MODULE, "cors") {
		c := cors.DefaultConfig()
		c.AllowAllOrigins = true
		c.AllowCredentials = true
		Engine.Use(cors.New(c))
	}

	//启用session
	Engine.Use(sessions.Sessions("iot-master", cookie.NewStore([]byte("iot-master"))))

	//开启压缩
	if config.GetBool(MODULE, "gzip") {
		Engine.Use(gzip.Gzip(gzip.DefaultCompression)) //gzip.WithExcludedPathsRegexs([]string{".*"})
	}

	JwtKey = config.GetString(MODULE, "jwt_key")
	JwtExpire = time.Hour * time.Duration(config.GetInt(MODULE, "jwt_expire"))

	//Engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return nil
}

func Shutdown() error {
	if Server == nil {
		return exception.New("服务未启动")
	}
	return Server.Shutdown(context.Background())
}

func _FileSystem2() *FileSystem {
	var fs FileSystem
	tm := time.Now()
	Engine.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			f, err := fs.Open(c.Request.URL.Path)
			if err == nil {
				defer f.Close()
				stat, err := f.Stat()
				if err != nil {
					c.Next() //500错误
					return
				}
				if !stat.IsDir() {
					fn := c.Request.URL.Path
					//fn := c.Request.URL.Path + ".html" //避免DetectContentType
					http.ServeContent(c.Writer, c.Request, fn, tm, f)
					return
				}
			}
		}
	})
	return &fs
}

func _RegisterFS(fs http.FileSystem, prefix, index string) {
	tm := time.Now()
	Engine.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			//支持前端框架的无“#”路由
			fn := path.Join(prefix, c.Request.URL.Path) //删除查询参数
			//fn := c.Request.URL.Path
			f, err := fs.Open(fn)
			if err == nil {
				defer f.Close()
				stat, err := f.Stat()
				if err != nil {
					c.Next() //500错误
					return
				}
				if !stat.IsDir() {
					http.ServeContent(c.Writer, c.Request, fn, tm, f)
					return
				}
			}

			//默认首页
			fn = path.Join(prefix, index) //删除查询参数
			f, err = fs.Open(fn)
			if err != nil {
				c.Next()
				return
			}
			defer f.Close()

			fn += ".html" //避免DetectContentType
			http.ServeContent(c.Writer, c.Request, fn, tm, f)
		}
	})
}

func Serve() error {

	//静态文件
	tm := time.Now()
	Engine.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			f, err := Static.Open(c.Request.URL.Path)
			if err == nil {
				defer f.Close()
				stat, err := f.Stat()
				if err != nil {
					c.Next() //500错误
					return
				}
				if !stat.IsDir() {
					fn := c.Request.URL.Path
					//fn := c.Request.URL.Path + ".html" //避免DetectContentType
					http.ServeContent(c.Writer, c.Request, fn, tm, f)
					return
				}
			}
		}
	})

	https := config.GetString(MODULE, "https")

	if https == "TLS" {
		return ServeTLS()
	} else if https == "LetsEncrypt" {
		return ServeLetsEncrypt()
	} else {
		return ServeHTTP()
	}
}

func ServeHTTP() error {
	port := config.GetInt(MODULE, "port")
	addr := ":" + strconv.Itoa(port)
	log.Info("Web Server", addr)
	//return Engine.Run(addr)
	Server = &http.Server{Addr: addr, Handler: Engine.Handler()}
	return Server.ListenAndServe()
}
