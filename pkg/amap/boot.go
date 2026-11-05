package amap

// amap 只是 HTTP 客户端工具（Geocode/Route/Weather/Cell 定位等），
// 无启动/关闭逻辑，不需要 boot.Register。
// 配置读取统一通过 config.go 里的 GetXxx() 从 viper 取。
