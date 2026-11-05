// Package api OEM 定制化配置公开接口
// 提供无需认证的 /api/oem 端点，供前端（App/Admin）获取品牌白标信息
package api

import (
	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/curd"

	"github.com/gin-gonic/gin"
)

// GetOEM 获取 OEM 定制化配置信息
// 路由路径：GET /api/oem（公开接口，无需认证）
func GetOEM(c *gin.Context) {
	curd.Success(c, gin.H{
		"logo":        config.GetString("oem", "logo"),
		"system_name": config.GetString("oem", "system_name"),
		"company":     config.GetString("oem", "company"),
		"icp":         config.GetString("oem", "icp"),
	})
}
