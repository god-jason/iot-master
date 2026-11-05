// Package api 认证辅助工具模块
// 提供从 gin.Context 获取当前登录用户信息的便捷函数，作为 auth.go 中基础函数的封装
package api

import (
	"github.com/gin-gonic/gin"
)

// GetCurrentUser 从 gin.Context 获取当前登录用户的完整 Claims 信息
// 路由: 无（工具函数）
// 参数: c - Gin 上下文，需先经过 JWTAuth 中间件写入 claims
// 返回: Claims 结构体指针；若未登录或解析失败则返回空的 Claims（各字段为零值）
// 说明: 封装了对 Context claims 的类型断言，调用方无需自行处理
func GetCurrentUser(c *gin.Context) *Claims {
	if claims, exists := c.Get("claims"); exists {
		if claim, ok := claims.(*Claims); ok {
			return claim
		}
	}
	return &Claims{}
}

// GetCurrentUserID 从 Context 获取当前登录用户的用户 ID
// 路由: 无（工具函数）
// 参数: c - Gin 上下文
// 返回: 用户 ID 字符串；未登录时返回空字符串
// 说明: 底层调用 GetUserID，是 GetUserID 的语义化别名
func GetCurrentUserID(c *gin.Context) string {
	return GetUserID(c)
}
