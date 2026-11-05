// Package api 认证与授权模块
// 提供 JWT Token 的生成与解析、API 认证中间件、权限校验以及用户基本信息查询等功能
package api

import (
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/logger"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenExpireHours 登录 token 默认有效期（小时）
const TokenExpireHours = 72

// errJWTSecretMissing JWT 签名密钥缺失或过短
var errJWTSecretMissing = errors.New("jwt.api 未配置或长度不足 16 位")

// jwtSecret 读取并校验 JWT 签名密钥，避免空密钥导致任意 token 可伪造
func jwtSecret() ([]byte, error) {
	secret := config.GetString("jwt", "api")
	if len(secret) < 16 {
		return nil, errJWTSecretMissing
	}
	return []byte(secret), nil
}

// Claims JWT 声明结构体
// 存储用户认证信息，在 API 请求间传递用户身份与权限标识
type Claims struct {
	UserID               string `json:"user_id"`            // 用户唯一标识
	Role                 int    `json:"role"`               // 用户角色编码（具体含义见业务字典）
	Admin                bool   `json:"admin"`              // 是否为平台管理员（true=管理员）
	Platform             bool   `json:"platform,omitempty"` // 是否平台账号（可登录管理后台）
	jwt.RegisteredClaims        // JWT 标准声明（过期时间、签发时间等）
}

// GetUserID 获取用户 ID
func (c *Claims) GetUserID() string {
	return c.UserID
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(userID string, role int, admin bool, platform bool, expireHours int) (string, error) {
	claims := Claims{
		UserID:   userID,
		Role:     role,
		Admin:    admin,
		Platform: platform,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析 JWT 令牌（强制 HS256，拒绝其它签名算法）
func ParseToken(tokenString string) (*Claims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UserID == "" {
		return nil, errors.New("无效的token")
	}
	return claims, nil
}

// JWTAuth API 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取 Authorization 请求头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少Authorization header"})
			c.Abort()
			return
		}

		// 拆分 "Bearer <token>" 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的Authorization格式"})
			c.Abort()
			return
		}

		// 解析并校验 JWT token
		claims, err := ParseToken(parts[1])
		if err != nil {
			logger.Warn("API token解析失败", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效或过期的token"})
			c.Abort()
			return
		}

		// 将解析出的用户信息写入 Context，供后续业务处理函数读取
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("admin", claims.Admin)
		c.Set("claims", claims)
		c.Next()
	}
}

// GetUserID 从 Context 获取当前登录用户的用户 ID
func GetUserID(c *gin.Context) string {
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			return id
		}
	}
	return ""
}
