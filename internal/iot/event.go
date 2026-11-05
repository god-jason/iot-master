package iot

import "github.com/god-jason/iot-master/pkg/event"

// bus 包级内部事件总线实例（MQTT 通配符 trie 实现），由 Publish/Subscribe 统一代理，
// 供各接入层与业务模块在同一进程内解耦通信。
var bus = event.NewEventBus()

// EventBus 返回内部事件总线实例（MQTT 通配符 trie 版）
func EventBus() event.EventBus {
	return bus
}

// Publish 异步发布事件
func Publish(topic string, payload any) {
	bus.Publish(topic, payload)
}

// Subscribe 订阅事件，返回取消订阅的函数
func Subscribe(filter string, callback event.Handler) func() {
	return bus.Subscribe(filter, callback)
}
