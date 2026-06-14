package api

// API 接口注册模块
// 提供接口注册、路由分组、鉴权等功能

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/web"
)

type API struct {
	Method   string                   // HTTP 方法
	Path     string                   // 路由路径
	Handlers []gin.HandlerFunc        // 处理函数
}

// 不用鉴权的接口
var apisUnauthorized []*API

// 接口
var apis []*API

// 管理员接口
var apisAdmin []*API

// RegisterUnAuthorized 注册无需鉴权的接口
func RegisterUnAuthorized(method, path string, handlers ...gin.HandlerFunc) {
	apisUnauthorized = append(apisUnauthorized, &API{method, path, handlers})
}

// Register 注册需要鉴权的接口
func Register(method, path string, handlers ...gin.HandlerFunc) {
	apis = append(apis, &API{method, path, handlers})
}

// RegisterAdmin 注册管理员接口
func RegisterAdmin(method, path string, handlers ...gin.HandlerFunc) {
	apisAdmin = append(apisAdmin, &API{method, path, handlers})
}

// catchError 捕获 panic 异常，防止服务崩溃
func catchError(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// runtime.Stack()
			// debug.Stack()
			switch err.(type) {
			case error:
				Error(ctx, err.(error))
			case string:
				Fail(ctx, err.(string))
			default:
				ctx.JSON(http.StatusOK, gin.H{"error": err})
			}
			// TODO 这里好像又继续了
		}
	}()
	ctx.Next()
	// TODO 内容如果为空，返回404
}

// mustLogin 验证用户登录状态
func mustLogin(ctx *gin.Context) {
	// 优先使用 query参数 token
	token := ctx.Request.URL.Query().Get("token")
	if token == "" {
		// 使用JWT
		token = ctx.Request.Header.Get("Authorization")
		if token != "" {
			// 此处需要去掉 Bearer
			if tkn, has := strings.CutPrefix(token, "Bearer "); has {
				token = tkn
			} else {
				// 错误
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "需要Bearer令牌"})
				ctx.Abort()
				return
			}
		}
	}

	// 验证JWT
	if token != "" {
		claims, err := web.JwtVerify(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			ctx.Abort()
		} else {
			ctx.Set("user", claims.ID) //与session统一
			ctx.Set("admin", claims.Admin)
			ctx.Set("tenant", claims.Tenant)
			ctx.Next()
		}
		return
	}

	// 检查Session
	session := sessions.Default(ctx)
	if user := session.Get("user"); user != nil {
		ctx.Set("user", user)
		ctx.Set("admin", session.Get("admin"))
		ctx.Set("tenant", session.Get("tenant"))
		ctx.Next()
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		ctx.Abort()
	}
}

// mustBeAdmin 验证管理员权限
func mustBeAdmin(ctx *gin.Context) {
	val, has := ctx.Get("admin")
	if !has {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "不是管理员"})
		ctx.Abort()
		return
	}
	if v, ok := val.(bool); !ok || !v {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "不是管理员"})
		ctx.Abort()
		return
	}
	ctx.Next()
}

// registerRoutes 注册路由组
func registerRoutes(base string) {
	// 默认api开头
	if base == "" {
		base = "api"
	}

	router := web.Engine().Group(base)

	// 错误恢复，并返回至前端
	router.Use(catchError)

	// 注册接口（不需要鉴权的）
	for _, a := range apisUnauthorized {
		router.Handle(a.Method, a.Path, a.Handlers...)
	}

	// 检查 session，必须登录
	router.Use(mustLogin)

	// 注册接口
	for _, a := range apis {
		router.Handle(a.Method, a.Path, a.Handlers...)
	}

	// 检查 session，必须登录
	router.Use(mustBeAdmin)

	// 注册接口
	for _, a := range apisAdmin {
		router.Handle(a.Method, a.Path, a.Handlers...)
	}

	// 附件管理
	// attach.Routers(router.Group("/attach"), "attach")

	// TODO 报接口错误（以下代码不生效，路由好像不是树形处理）
	router.Use(func(ctx *gin.Context) {
		Fail(ctx, "接口不存在")
		ctx.Abort()
	})
}
