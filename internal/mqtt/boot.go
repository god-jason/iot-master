package mqtt

import "github.com/god-jason/iot-master/pkg/boot"

func init() {
	boot.Register(MODULE, &boot.Task{
		Startup:  BootStartup,
		Shutdown: Shutdown,
		Depends:  []string{"pool", "config"},
	})
}

// BootStartup boot 启动入口：连接 MQTT broker 后启动上行订阅
func BootStartup() error {
	if err := Startup(); err != nil {
		return err
	}
	if !Enabled() {
		return nil // enable=false 时没有连接对象，跳过订阅
	}
	StartSubscribe()
	return nil
}
