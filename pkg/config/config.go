// Package config 基于 viper 的全局配置中心。
// 所有 pkg 模块在 init() 里调用 config.SetDefault(module, key, value) 注册默认值，
// boot.Startup() 自动加载 master.yaml 覆盖默认值，环境变量 DMP_MODULE_KEY 也能覆盖。
// 消费方通过 config.GetString/GetBool/GetInt/GetStringSlice 直取，不需要中间 struct。
package config

import (
	"github.com/spf13/viper"

	"github.com/god-jason/iot-master/pkg/boot"
)

// init 里注册 boot 任务（但 viper 的目录/文件名需要 Startup 前由 Name/AddPath 配置）
func init() {
	boot.Register("config", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
	})

	// 环境变量：前缀 DMP_，自动绑定
	viper.SetEnvPrefix("dmp")
	viper.AutomaticEnv()

	// 杂项配置默认值（不属于任何独立 pkg 模块）
	SetDefault("jwt", "api", "iot-dmp-secret-key-change-me")
	SetDefault("upload", "dir", "./uploads")
	SetDefault("upload", "server_url", "")
	SetDefault("oem", "system_name", "IoT DMP")
	SetDefault("oem", "company", "")
	SetDefault("oem", "logo", "")
	SetDefault("oem", "icp", "")
}

// Name 设置配置文件名称（不含扩展名），如 "config"
func Name(name string) {
	viper.SetConfigName(name)
}

// AddPath 追加一个配置搜索目录
func AddPath(path string) {
	viper.AddConfigPath(path)
}

// SetType 设置配置文件类型（默认 yaml）
func SetType(t string) {
	viper.SetConfigType(t)
}

// Load 加载配置文件。auto=true 时文件不存在会自动创建一份。
func Load(auto bool) error {
	if err := viper.ReadInConfig(); err != nil {
		if auto {
			return viper.SafeWriteConfig()
		}
		return err
	}
	return nil
}

// Store 把当前配置写回配置文件
func Store() error {
	return viper.WriteConfig()
}

// SetDefault 注册模块默认值（module.key）
func SetDefault(module string, key string, value any) {
	viper.SetDefault(module+"."+key, value)
}

// Set 运行时覆盖某个配置
func Set(module string, key string, value any) {
	viper.Set(module+"."+key, value)
}

// GetString 读字符串
func GetString(module string, key string) string {
	return viper.GetString(module + "." + key)
}

// GetBool 读布尔
func GetBool(module string, key string) bool {
	return viper.GetBool(module + "." + key)
}

// GetInt 读整数
func GetInt(module string, key string) int {
	return viper.GetInt(module + "." + key)
}

// GetFloat 读浮点
func GetFloat(module string, key string) float64 {
	return viper.GetFloat64(module + "." + key)
}

// GetStringSlice 读字符串数组
func GetStringSlice(module string, key string) []string {
	return viper.GetStringSlice(module + "." + key)
}

// IsSet 判断某配置项是否存在
func IsSet(module string, key string) bool {
	return viper.IsSet(module + "." + key)
}

// Unmarshal 把整个 viper 配置反序列化到结构体
func Unmarshal(v any) error {
	return viper.Unmarshal(v)
}

// Viper 返回底层 viper 实例（需要高级操作时用）
func Viper() *viper.Viper {
	return viper.GetViper()
}

// Startup boot 启动入口：加载 master.yaml（auto=true 自动生成默认配置）
func Startup() error {
	Name("master")
	AddPath(".")
	return Load(true)
}

// Shutdown boot 关闭入口
func Shutdown() error {
	return nil
}
