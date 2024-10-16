package protocol

import (
	"github.com/god-jason/bucket/types"
)

// Adapter 设备驱动
type Adapter interface {

	//Mount 挂载设备
	Mount(deviceId string, productId string, station types.Options) error

	//Unmount 卸载设备
	Unmount(deviceId string) error

	//Get 读数据
	Get(deviceId, point string) (any, error)

	//Set 写数据
	Set(deviceId, point string, value any) error

	//Poll 读取所有数据
	Poll(deviceId string) (map[string]any, error)
}
