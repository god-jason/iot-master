package config

import (
	"github.com/busy-cloud/boat/lib"
	"github.com/spf13/viper"
)

func init() {
	//引入viper配置文件
	//viper.SetConfigName("config") //name := lib.AppName()
	viper.SetConfigName(lib.AppName())
	viper.SetConfigType("yaml")
	viper.AddConfigPath(lib.AppDir())
	viper.AddConfigPath(".")

	//绑定命令行参数
	//_ = viper.BindPFlags(pflag.CommandLine)

	//绑定环境变量
	viper.SetEnvPrefix("boat")
	viper.AutomaticEnv()
}

func Name(name string) {
	viper.SetConfigName(name)
}

func Load(auto bool) error {
	err := viper.ReadInConfig()
	if err != nil {
		if auto {
			//自动创建
			return viper.SafeWriteConfig()
		} else {
			return err
		}
	}
	return nil
}

func Store() error {
	return viper.WriteConfig()
}

func SetDefault(module string, key string, value any) {
	viper.SetDefault(module+"."+key, value)
}

func Set(module string, key string, value any) {
	viper.Set(module+"."+key, value)
}

func GetBool(module string, key string) bool {
	return viper.GetBool(module + "." + key)
}

func GetString(module string, key string) string {
	return viper.GetString(module + "." + key)
}

func GetInt(module string, key string) int {
	return viper.GetInt(module + "." + key)
}

func GetFloat(module string, key string) float64 {
	return viper.GetFloat64(module + "." + key)
}

func GetStringSlice(module string, key string) []string {
	return viper.GetStringSlice(module + "." + key)
}
