// Package ota OTA 固件升级：目标解析、灰度批次调度、MQTT 下发、进度上报与超时重试。
package ota

import (
	"fmt"
	"strings"
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/config"
)

// MODULE OTA 配置模块名
const MODULE = "ota"

func init() {
	config.SetDefault(MODULE, "enabled", true)
	config.SetDefault(MODULE, "storage_dir", "./uploads/firmware")
	config.SetDefault(MODULE, "server_url", "")
	config.SetDefault(MODULE, "max_size_mb", 200)
	config.SetDefault(MODULE, "concurrency", 20)
	config.SetDefault(MODULE, "scan_interval", "10s")
	config.SetDefault(MODULE, "ack_timeout", "30s")
	config.SetDefault(MODULE, "report_timeout_minutes", 10)
	config.SetDefault(MODULE, "max_retry", 2)
	config.SetDefault(MODULE, "default_success_threshold", 0.95)
	config.SetDefault(MODULE, "default_failure_threshold", 0.10)
}

// Enabled OTA 模块是否启用
func Enabled() bool { return config.GetBool(MODULE, "enabled") }

// StorageDir 固件存储目录
func StorageDir() string {
	dir := config.GetString(MODULE, "storage_dir")
	if dir == "" {
		dir = "./uploads/firmware"
	}
	return dir
}

// MaxSizeBytes 单个固件大小上限（字节）
func MaxSizeBytes() int64 {
	mb := config.GetInt(MODULE, "max_size_mb")
	if mb <= 0 {
		mb = 200
	}
	return int64(mb) * 1024 * 1024
}

// ServerBase 设备可访问的平台基础地址。
// 优先 ota.server_url，其次复用 upload.server_url；都未配置时退回本机 web 端口。
// 注意：下载地址绝不能依赖请求 Host，否则设备与后台网络不通时无法下载。
func ServerBase() string {
	if v := strings.TrimSpace(config.GetString(MODULE, "server_url")); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	if v := strings.TrimSpace(config.GetString("upload", "server_url")); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	if port := config.GetInt("web", "port"); port > 0 {
		return fmt.Sprintf("http://127.0.0.1:%d", port)
	}
	return ""
}

// PublicFirmwareURL 固件下载相对路径（无需 token）
func PublicFirmwareURL(id string) string {
	return "/api/ota/firmware/" + id + "/file"
}

// DownloadURL 把固件相对下载路径拼成设备可访问的完整地址
func DownloadURL(fw *types.Firmware) string {
	if fw == nil {
		return ""
	}
	path := fw.FileURL
	if path == "" {
		path = PublicFirmwareURL(fw.ID)
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	base := ServerBase()
	if base == "" {
		return path
	}
	return base + path
}

// ScanInterval 调度扫描间隔
func ScanInterval() time.Duration { return parseDuration("scan_interval", 10*time.Second) }

// AckTimeout 下发后等待设备 ACK 的超时（传输层自身亦有超时，这里用于文档与重试判断）
func AckTimeout() time.Duration { return parseDuration("ack_timeout", 30*time.Second) }

// ReportTimeout 设备无进度上报的超时时间
func ReportTimeout() time.Duration {
	minutes := config.GetInt(MODULE, "report_timeout_minutes")
	if minutes <= 0 {
		minutes = 10
	}
	return time.Duration(minutes) * time.Minute
}

// Concurrency 同时下发设备的并发数
func Concurrency() int {
	n := config.GetInt(MODULE, "concurrency")
	if n <= 0 {
		n = 20
	}
	return n
}

// MaxRetry 单设备最大重试次数
func MaxRetry() int {
	n := config.GetInt(MODULE, "max_retry")
	if n < 0 {
		n = 0
	}
	return n
}

// DefaultSuccessThreshold 默认批次成功率门禁
func DefaultSuccessThreshold() float64 {
	v := config.GetFloat(MODULE, "default_success_threshold")
	if v <= 0 {
		v = 0.95
	}
	return v
}

// DefaultFailureThreshold 默认批次失败率上限
func DefaultFailureThreshold() float64 {
	v := config.GetFloat(MODULE, "default_failure_threshold")
	if v < 0 {
		v = 0.10
	}
	return v
}

// parseDuration 支持 "10s" 与整数秒两种配置写法
func parseDuration(key string, def time.Duration) time.Duration {
	if s := config.GetString(MODULE, key); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			return d
		}
	}
	if n := config.GetInt(MODULE, key); n > 0 {
		return time.Duration(n) * time.Second
	}
	return def
}
