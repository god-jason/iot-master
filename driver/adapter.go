package driver

import (
	"github.com/god-jason/iot-master/types"
)

// Adapter 驱动接口
type Adapter interface {

	//Mount 挂载设备
	Mount(station types.Options, product string) (device Device, err error)
}

// Device 设备实例接口
type Device interface {

	//Set 设置属性值
	Set(name string, value any) error

	//Get 获得属性值
	Get(name string) (value any, err error)

	//GetAll 获得所有属性值
	GetAll() (map[string]any, error)

	//Do 执行动作
	Do(name string, params map[string]any) (returns map[string]any, err error)

	//Unmount 卸载设备
	Unmount() error
}
