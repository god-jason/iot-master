package weixin

import (
	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/god-jason/iot-master/pkg/boot"
	"github.com/god-jason/iot-master/pkg/config"
)

func init() {
	boot.Register("weixin", &boot.Task{
		Startup:  startup,
		Shutdown: nil,
		Depends:  nil,
	})
}

var mp *miniProgram.MiniProgram

func startup() (err error) {
	// 初始化微信小程序配置
	cfg := miniProgram.UserConfig{
		AppID:  config.GetString(MODULE, "appid"),
		Secret: config.GetString(MODULE, "secret"),
	}

	mp, err = miniProgram.NewMiniProgram(&cfg)
	if err != nil {
		return err
	}

	return nil
}
