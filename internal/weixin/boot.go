package weixin

// weixin SDK 全局单例初始化由调用方显式调用 InitMiniProgram() 完成，
// 无 boot 启动逻辑（微信 SDK 不需要连接/轮询/后台 goroutine 管理）。
