package web

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/logger"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Recovery 全局 panic recover 中间件 — 捕获 handler 内的 panic，
// 打印完整堆栈到 logger + 500 JSON 响应，防止 goroutine 崩溃杀掉整个进程
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				logger.Error("[RECOVERY] 捕获到 panic",
					"error", fmt.Sprintf("%v", err),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"query", c.Request.URL.RawQuery,
					"ip", c.ClientIP(),
					"stack", stack,
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "服务器内部错误，请查看日志",
				})
			}
		}()
		c.Next()
	}
}

// ============ 限流器 ============

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]int64
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]int64),
		limit:    limit,
		window:   window,
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()
	windowStart := now - int64(r.window.Seconds())

	var validRequests []int64
	for _, ts := range r.requests[key] {
		if ts > windowStart {
			validRequests = append(validRequests, ts)
		}
	}
	r.requests[key] = validRequests

	if len(validRequests) >= r.limit {
		return false
	}

	r.requests[key] = append(validRequests, now)
	return true
}

// RateLimit 限流中间件（基于 user_id string 或 IP）
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		var key string
		if userID, exists := c.Get("user_id"); exists {
			if str, ok := userID.(string); ok {
				key = str
			}
		}
		if key == "" {
			key = c.ClientIP()
		}

		if !limiter.Allow(key) {
			c.JSON(http.StatusOK, gin.H{"code": 429, "message": "请求过于频繁，请稍后再试"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============ ID解析中间件 ============

// ParseStringID 将路径中的 string ID 存入 context
func ParseStringID(paramName ...string) gin.HandlerFunc {
	name := "id"
	if len(paramName) > 0 && paramName[0] != "" {
		name = paramName[0]
	}

	return func(c *gin.Context) {
		id := c.Param(name)
		if id == "" {
			c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID格式"})
			c.Abort()
			return
		}
		c.Set(name, id)
		c.Next()
	}
}
