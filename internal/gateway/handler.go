package gateway

import (
	"context"
)

// ProtocolHandler 业务回调接口
//
// 所有回调在 gnet event loop 外异步调用，可以安全访问 DB / 调用 engine。
type ProtocolHandler interface {
	// OnRegister 链路建立（注册包解析成功）：业务层按 LinkID 加载设备并 Mount
	OnRegister(ctx context.Context, link *Link)

	// OnDisconnect 链路断开
	OnDisconnect(ctx context.Context, link *Link, reason string)
}

// defaultHandler ProtocolHandler 的空实现
//
// 未通过 WithHandler 注入业务回调时使用：只完成协议层的连接注册，
// 不加载设备、不注册 iot，适合仅观察链路或测试场景。
type defaultHandler struct{}

// OnRegister 空实现：不加载设备，链路仅停留在协议注册状态
func (defaultHandler) OnRegister(context.Context, *Link) {}

// OnDisconnect 空实现：不影响任何设备，无需清理 iot
func (defaultHandler) OnDisconnect(context.Context, *Link, string) {}
