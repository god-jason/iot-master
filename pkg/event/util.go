package event

import (
	"reflect"
	"unsafe"
)

// safeCall 带 recover 的 handler 调用，保证单个 handler panic 不影响其他订阅者
func safeCall(h Handler, topic string, payload any) {
	defer func() { _ = recover() }()
	h(topic, payload)
}

// handlerKey 把 Handler 转成 uintptr 用于 map 去重 key
// 函数值本身是比较类型（相同的函数值底层指向同一个函数地址），
// 但 Go 不允许函数值做 map key — 这里用 reflect 取出其底层指针。
func handlerKey(h Handler) uintptr {
	if h == nil {
		return 0
	}
	v := reflect.ValueOf(h)
	return uintptr(unsafe.Pointer(v.Pointer()))
}
