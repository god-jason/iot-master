package amap

import (
	"github.com/god-jason/iot-master/pkg/config"
)

// MODULE 模块名称，用于配置管理
const MODULE = "amap"

func init() {
	// 默认服务器地址，高德地图Web服务API地址
	config.SetDefault(MODULE, "server", "https://restapi.amap.com")
	// 默认API密钥，需要用户自行配置
	config.SetDefault(MODULE, "key", "")
	// V1基站定位API地址（旧版）
	config.SetDefault(MODULE, "apilocate_url", "https://apilocate.amap.com/position")
	// V2基站定位API地址（新版IoT定位）
	config.SetDefault(MODULE, "cell_url", "https://restapi.amap.com/v5/position/IoT")
}

// GetKey 获取配置的高德地图API密钥
func GetKey() string {
	return config.GetString(MODULE, "key")
}

// SetKey 设置高德地图API密钥
func SetKey(key string) {
	config.Set(MODULE, "key", key)
}

// GetServer 获取配置的服务器地址
func GetServer() string {
	return config.GetString(MODULE, "server")
}

// SetServer 设置服务器地址
func SetServer(server string) {
	config.Set(MODULE, "server", server)
}

// GetApilocateURL 获取V1基站定位API地址
func GetApilocateURL() string {
	return config.GetString(MODULE, "apilocate_url")
}

// SetApilocateURL 设置V1基站定位API地址
func SetApilocateURL(url string) {
	config.Set(MODULE, "apilocate_url", url)
}

// GetCellURL 获取V2基站定位API地址
func GetCellURL() string {
	return config.GetString(MODULE, "cell_url")
}

// SetCellURL 设置V2基站定位API地址
func SetCellURL(cellURL string) {
	config.Set(MODULE, "cell_url", cellURL)
}
