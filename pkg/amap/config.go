package amap

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper 模块名
const MODULE = "amap"

func init() {
	config.SetDefault(MODULE, "key", "")
	config.SetDefault(MODULE, "server", "https://restapi.amap.com")
	config.SetDefault(MODULE, "apilocate_url", "https://apilocate.amap.com/position")
	config.SetDefault(MODULE, "cell_url", "https://restapi.amap.com/v5/position/IoT")
}

// GetKey 获取高德地图 API Key
func GetKey() string {
	return config.GetString(MODULE, "key")
}

// GetServer 获取高德地图 Web 服务地址
func GetServer() string {
	return config.GetString(MODULE, "server")
}

// GetApilocateURL 获取 V1 基站定位 API 地址
func GetApilocateURL() string {
	return config.GetString(MODULE, "apilocate_url")
}

// GetCellURL 获取 V2 基站定位 API 地址
func GetCellURL() string {
	return config.GetString(MODULE, "cell_url")
}
