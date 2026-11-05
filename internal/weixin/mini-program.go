// Package weixin 微信小程序登录与手机号获取相关能力
// 基于 PowerWeChat SDK 封装 wx.login 凭证校验与 getPhoneNumber 手机号获取。
package weixin

import (
	"context"
	"fmt"

	"github.com/god-jason/iot-master/pkg/config"
)

// MiniProgramSession 微信小程序登录会话结构
// 对应微信接口 jscode2session 的返回字段
type MiniProgramSession struct {
	OpenID     string `json:"openid"`      // 用户唯一标识（当前小程序范围内唯一）
	SessionKey string `json:"session_key"` // 会话密钥，用于加密数据解密（不建议暴露给前端）
	UnionID    string `json:"unionid"`     // 用户在开放平台的统一标识（同一开发者多应用间一致）
}

// CodeToSession 通过 wx.login 返回的 code 换取微信小程序登录会话
// 路由路径：POST /api/weixin/code2session（HTTP 层路由注册在 api 包 registerPublicRoutes）
//
// 处理逻辑：
//  1. 未配置小程序 AppID/Secret 时直接报错，禁止使用 mock 身份
//  2. 调用 PowerWeChat SDK 的 Auth.Session 接口向微信服务器换取 openid / session_key / unionid
//
// 参数:
//   - code: wx.login 接口返回的临时登录凭证，有效期 5 分钟
//
// 返回:
//   - *MiniProgramSession: 登录会话信息
//   - error: 小程序未初始化 / 微信接口调用失败时返回错误
func CodeToSession(code string) (*MiniProgramSession, error) {
	appID := config.GetString("weixin", "mini_program.app_id")
	appSecret := config.GetString("weixin", "mini_program.app_secret")
	if appID == "" || appSecret == "" {
		// 未配置小程序凭证时直接报错，禁止使用 mock 身份（避免生产误配被绕过登录）
		return nil, fmt.Errorf("微信小程序未配置 app_id/app_secret")
	}

	// 校验 SDK 全局实例已初始化
	if MiniProgramApp == nil {
		return nil, fmt.Errorf("微信小程序未初始化")
	}

	// 调用微信 jscode2session 接口
	result, err := MiniProgramApp.Auth.Session(context.Background(), code)
	if err != nil {
		return nil, err
	}

	return &MiniProgramSession{
		OpenID:     result.OpenID,
		SessionKey: result.SessionKey,
		UnionID:    result.UnionID,
	}, nil
}

// GetUserPhoneNumber 通过 getPhoneNumber 组件返回的 code 换取用户手机号
// 路由路径：POST /api/weixin/phone（HTTP 层路由注册在 api 包 registerPublicRoutes）
//
// 处理逻辑：
//  1. 未配置小程序凭证时直接报错，不返回测试手机号
//  2. 调用 PowerWeChat SDK 的 PhoneNumber.GetUserPhoneNumber 接口
//  3. 校验返回的 ErrCode 与 PhoneInfo 非空
//
// 参数:
//   - code: getPhoneNumber 按钮回调中获取的手机号凭证 code
//
// 返回:
//   - string: 用户纯文本手机号（如 138xxxxxxxx）
//   - error: 小程序未初始化 / 微信接口失败 / 手机号信息为空时返回错误
func GetUserPhoneNumber(code string) (string, error) {
	appID := config.GetString("weixin", "mini_program.app_id")
	appSecret := config.GetString("weixin", "mini_program.app_secret")
	if appID == "" || appSecret == "" {
		// 未配置小程序凭证时直接报错，禁止返回固定测试手机号
		return "", fmt.Errorf("微信小程序未配置 app_id/app_secret")
	}

	// 校验 SDK 全局实例已初始化
	if MiniProgramApp == nil {
		return "", fmt.Errorf("微信小程序未初始化")
	}

	// 调用微信手机号获取接口
	result, err := MiniProgramApp.PhoneNumber.GetUserPhoneNumber(context.Background(), code)
	if err != nil {
		return "", err
	}

	// 微信侧返回业务错误码（非 0 表示失败）
	if result.ErrCode != 0 {
		return "", fmt.Errorf("获取手机号失败: %s", result.ErrMsg)
	}

	// 兜底校验手机号信息非空
	if result.PhoneInfo == nil {
		return "", fmt.Errorf("手机号信息为空")
	}

	return result.PhoneInfo.PhoneNumber, nil
}
