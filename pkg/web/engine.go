package web

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/god-jason/iot-master/pkg/config"

	"github.com/gin-gonic/gin"
)

type Engine struct {
	*gin.Engine
}

// New 创建新的Web引擎
func New() *Engine {
	if config.GetBool(MODULE, "debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(Recovery())
	engine.Use(CORS())
	//engine.Use(gzip.Gzip(gzip.DefaultCompression))

	e := &Engine{Engine: engine}

	// 服务前端静态页面（目录不存在时跳过，不影响后端 API）
	if staticDir := config.GetString(MODULE, "static"); staticDir != "" {
		if err := e.ServeFrontend(staticDir); err != nil {
			// 静态目录不存在时静默跳过（开发环境前端独立启动时不需要后端服务静态页面）
			// 其他错误（如权限）才 panic
			if os.IsNotExist(err) {
				// silent skip
			} else {
				panic("服务前端静态页面失败: " + err.Error())
			}
		}
	}

	return e
}

// NewBasic 创建基础引擎（不包含默认中间件）
func NewBasic() *Engine {
	if config.GetBool(MODULE, "debug") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return &Engine{Engine: gin.New()}
}

// ServeFrontend 服务前端静态页面（支持 SPA 路由）
// 静态文件优先匹配，未命中的路径返回 index.html 由前端路由处理
func (e *Engine) ServeFrontend(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// 确保目录存在
	if _, err := os.Stat(absDir); err != nil {
		return err
	}

	e.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 等后端路由未命中时返回 JSON 404
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/open/") || strings.HasPrefix(path, "/node/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}

		// 尝试服务静态文件
		filePath := filepath.Join(absDir, path)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			c.File(filePath)
			return
		}

		// SPA 路由：返回 index.html
		c.File(filepath.Join(absDir, "index.html"))
	})

	return nil
}
