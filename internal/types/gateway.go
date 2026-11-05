package types

import (
	"time"
)

// GatewayDevice 网关默认设备配置（内嵌为 JSON，不落独立表）
type GatewayDevice struct {
	Name    string `json:"name"`     // 设备名称
	Address string `json:"address"`  // 设备地址（从站号 / 编号等）
	ModelID string `json:"model_id"` // 设备使用的物模型ID
}

// GatewayServer Gateway TCP 服务器实例配置表
// 系统可同时启动多个 Gateway 服务器，每个监听不同端口、独立调度
type GatewayServer struct {
	ID             int64           `json:"id" xorm:"id bigint pk autoincr"`                         // 服务器ID
	Name           string          `json:"name" xorm:"name varchar(64) notnull"`                    // 服务器名称
	Host           string          `json:"host" xorm:"host varchar(64)"`                            // 监听地址
	Port           int             `json:"port" xorm:"port int notnull"`                            // 监听端口
	Protocol       string          `json:"protocol" xorm:"protocol varchar(16)"`                    // tcp / rtu
	Interval       string          `json:"interval" xorm:"interval varchar(16)"`                    // 默认轮询间隔，如 "5s"
	Concurrency    int             `json:"concurrency" xorm:"concurrency int default 4"`            // 并发轮询 worker 数
	DefaultDevices []GatewayDevice `json:"default_devices" xorm:"default_devices json"`             // 默认设备列表，用于未指定设备的场景
	Disabled       bool            `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Created        time.Time       `json:"created,omitempty" xorm:"created datetime"`               // 创建时间
	Updated        time.Time       `json:"updated,omitempty" xorm:"updated datetime"`               // 更新时间
}

// TableName 显式指定表名，GatewayServer 落库到 gateway_server 表
func (GatewayServer) TableName() string { return "gateway_server" }

// Link 软网关链路 —— 一条注册成功的 TCP 连接（通常对应一台 DTU / 网关）。
//
// 设备通过 Device.LinkID 归属到 Link：Link 建立后按 link_id 查询其下设备，
// 并逐一 Mount 到该连接的协议实例。ID 即注册包中的 SN（无 SN 时取 IMEI）。
type Link struct {
	ID       string    `json:"id" xorm:"id char(32) pk notnull"`                        // 链路ID（注册包 SN，无 SN 时取 IMEI）
	ServerID int64     `json:"server_id" xorm:"server_id bigint index notnull"`         // 所属 Gateway 服务器ID → GatewayServer(int64)
	Name     string    `json:"name" xorm:"name varchar(128)"`                           // 链路名称
	Protocol string    `json:"protocol" xorm:"protocol varchar(16)"`                    // 链路自定义协议，为空则沿用所属服务器配置
	Disabled bool      `json:"disabled,omitempty" xorm:"disabled tinyint(1) default 0"` // 是否禁用
	Created  time.Time `json:"created,omitempty" xorm:"created datetime"`               // 创建时间
	Updated  time.Time `json:"updated,omitempty" xorm:"updated datetime"`               // 更新时间
}

// TableName 显式指定表名，Link 落库到 link 表
func (Link) TableName() string { return "link" }
