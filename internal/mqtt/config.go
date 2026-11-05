package mqtt

import "github.com/god-jason/iot-master/pkg/config"

// MODULE viper/boot 模块名
const MODULE = "mqtt"

func init() {
	// enable 是否启用 MQTT 客户端（默认关闭）
	config.SetDefault(MODULE, "enable", false)
	// url broker 地址，如 tcp://127.0.0.1:1883；为空表示不主动连 broker（内嵌直通场景）
	config.SetDefault(MODULE, "url", "")
	// client_id 客户端 ID，为空时启动阶段自动生成 12 位随机串
	config.SetDefault(MODULE, "client_id", "")
	// username 连接 broker 的用户名
	config.SetDefault(MODULE, "username", "")
	// password 连接 broker 的密码
	config.SetDefault(MODULE, "password", "")
}

// Enabled 是否启用 MQTT 客户端
func Enabled() bool { return config.GetBool(MODULE, "enable") }
