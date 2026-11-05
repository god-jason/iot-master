// Package open 封装开放 HTTP 接口：
// 离线场景执行 API（局域网客户端直连网关查询并执行场景，断 master 时仍可用）。
//
// 场景函数变量（GetEnabledScenesFn / ExecuteSceneFn）由调用方注入，打破 open ↔ 业务包循环依赖。
package open

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/god-jason/iot-master/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetEnabledScenesFn 返回当前启用的场景列表，由调用方注入
// ExecuteSceneFn 执行指定场景并返回逐动作结果，由调用方注入
var (
	GetEnabledScenesFn func() []types.Scene
	ExecuteSceneFn     func(sceneID int64) ([]types.SceneActionResult, error)
)

// RegisterRoutesWithJWT 带 JWT 鉴权的路由注册
func RegisterRoutesWithJWT(r *gin.RouterGroup, secret string) {
	api := r.Group("")
	api.Use(jwtAuth(secret))
	{
		api.GET("/scene", listScenes)
		api.POST("/scene/:id/execute", executeScene)
	}
}

// jwtAuth 构造开放接口的 JWT 鉴权中间件（仅校验 token 有效性，不解析业务身份）
func jwtAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Authorization header is required"})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Invalid authorization header format"})
			c.Abort()
			return
		}
		token, err := jwt.ParseWithClaims(parts[1], &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Invalid or expired token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// —— 离线场景 ——

// listScenes 返回当前启用的场景列表（供断网时的客户端展示与选择）
func listScenes(c *gin.Context) {
	if GetEnabledScenesFn == nil {
		c.JSON(503, gin.H{"code": 503, "message": "场景引擎未就绪"})
		return
	}
	scenes := GetEnabledScenesFn()
	list := make([]gin.H, 0, len(scenes))
	for _, s := range scenes {
		list = append(list, gin.H{"id": s.ID, "name": s.Name, "icon": s.Icon, "desc": s.Desc, "actions": s.Actions})
	}
	c.JSON(200, gin.H{"code": 0, "data": list})
}

// executeScene 按场景 ID 立即执行场景，返回逐动作的执行结果
func executeScene(c *gin.Context) {
	id64, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的场景ID"})
		return
	}
	if ExecuteSceneFn == nil {
		c.JSON(503, gin.H{"code": 503, "message": "场景引擎未就绪"})
		return
	}
	results, err := ExecuteSceneFn(id64)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "data": gin.H{"results": results}})
}
