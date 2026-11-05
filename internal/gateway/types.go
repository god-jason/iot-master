package gateway

import (
	"time"
)

// RegisterPacket 首包注册结构（JSON 模式）
//
// 注册包支持两种方式：
//   - 纯文本：整行内容即设备 SN（如 "GW0001\r\n"）
//   - JSON：  {"sn":"GW0001","imei":"...","model":"..."}，以换行或紧跟的协议帧分隔
type RegisterPacket struct {
	SN     string `json:"sn,omitempty"`     // 设备序列号，首选链路 ID
	IMEI   string `json:"imei,omitempty"`   // 模组 IMEI，SN 缺失时作为链路 ID
	Device string `json:"device,omitempty"` // 设备标识，SN/IMEI 均缺失时兜底
	Name   string `json:"name,omitempty"`   // 链路/设备名称
	Model  string `json:"model,omitempty"`  // 物模型 ID（可选）
}

// PollConfig 兜底轮询配置 —— 仅 Interval 生效，点位由挂载的子设备物模型决定
type PollConfig struct {
	Interval time.Duration // 采集间隔，<=0 时回落到 Server 默认间隔
}
