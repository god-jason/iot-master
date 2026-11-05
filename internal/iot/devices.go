package iot

import (
	"github.com/god-jason/iot-master/pkg/lib"
)

// devices 设备运行时 devices（所有接入类型手动注册）。
//
// broker / mqtt / gateway 在监测到设备接入时，构造对应的 Device 实现并 Register 进来；
// 设备断开时 Unregister。三种实现共享同一个 devices，业务层通过 GetDevice 统一查询。
var devices = lib.NewMap[string, Device]()

// Register 注册或覆盖设备（同一设备 ID 重复注册时后者覆盖前者）
func Register(deviceID string, dev Device) {
	devices.Set(deviceID, dev)
}

// GetDevice 获取设备（返回 nil,false 表示未注册）
func GetDevice(deviceID string) (Device, bool) {
	return devices.Get(deviceID)
}

// Unregister 注销设备
func Unregister(deviceID string) {
	devices.Delete(deviceID)
}

// Devices 返回所有已注册设备的拷贝，调用方可安全读写，不会篡改内部 registry
func Devices() *lib.Map[string, Device] {
	cp := lib.NewMap[string, Device]()
	devices.Range(func(deviceID string, dev Device) bool {
		cp.Set(deviceID, dev)
		return true
	})
	return cp
}

// DeviceKeys 返回所有已注册设备 ID
func DeviceKeys() []string {
	return devices.Keys()
}

// DeviceLen 已注册设备数量
func DeviceLen() int {
	return devices.Len()
}

// ClearDevice 清空全部注册（boot 关闭时调用）
func ClearDevice() {
	devices.Clear()
}
