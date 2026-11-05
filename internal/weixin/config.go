package weixin

import (
	"github.com/god-jason/iot-master/pkg/config"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
)

// MODULE viper 模块名
const MODULE = "weixin"

func init() {
	// 默认不配置小程序凭证：AppID / AppSecret 留空表示关闭微信相关能力
	config.SetDefault(MODULE, "mini_program.app_id", "")
	config.SetDefault(MODULE, "mini_program.app_secret", "")
}

// MiniProgramApp 微信小程序 SDK 全局单例，未配置凭证时为 nil
var MiniProgramApp *miniProgram.MiniProgram

// InitMiniProgram 初始化微信小程序 SDK（从 viper 读 weixin.mini_program.app_id/app_secret）
// 任一为空则跳过初始化
func InitMiniProgram() {
	appID := config.GetString(MODULE, "mini_program.app_id")
	appSecret := config.GetString(MODULE, "mini_program.app_secret")

	// 凭证不完整时保持单例为 nil，调用方据此走 mock 或返回未初始化错误
	if appID == "" || appSecret == "" {
		return
	}

	MiniProgramApp, _ = miniProgram.NewMiniProgram(&miniProgram.UserConfig{
		AppID:  appID,
		Secret: appSecret,
	})
}
