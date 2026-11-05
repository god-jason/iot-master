// Package thing 物模型定义（纯数据结构，无 DB tag / 无业务逻辑）
//
// 物模型描述设备的完整能力：属性 / 事件 / 服务 / 配置 / 控件
// 这些结构会被 JSON 序列化后存入 DB 的 properties/events/services 等 json 列。
//
// 注意：internal/types.Model 有 xorm tag（DB 模型），
// 其中嵌套的 Properties/Events/Services 字段就是本包的类型。
package thing

import "github.com/expr-lang/expr/vm"

// Property 属性定义，描述设备的可监控数据项
type Property struct {
	Key       string `json:"key,omitempty"`       // 唯一标识（如：temperature）
	Name      string `json:"name,omitempty"`      // 名称（如：温度）
	Type      string `json:"type,omitempty"`      // 数据类型：int / float / string / bool / enum
	Unit      string `json:"unit,omitempty"`      // 单位（如：℃）
	Precision int    `json:"precision,omitempty"` // 精度（小数位数）
	Mode      string `json:"mode,omitempty"`      // 访问模式：r / w / rw

	Threshold float64 `json:"threshold,omitempty"` // 阈值（告警用）
	Min       float64 `json:"min,omitempty"`        // 最小值（max > min 时生效，MQTT 上传范围过滤）
	Max       float64 `json:"max,omitempty"`        // 最大值（max > min 时生效）

	// 协议扩展字段（网关协议层使用）
	Register     string  `json:"register,omitempty"`      //寄存器类型，比如 holding_register / input_register / discrete_input / coil
	Address      string  `json:"address,omitempty"`       // 寄存器地址（支持三种格式，比如：1 0x0001 H0001）
	Format       string  `json:"format,omitempty"`        // uint8 / int8 / uint16 / int16 / uint32 / int32 / float32 / float64 / bcd 等
	LittleEndian bool    `json:"little_endian,omitempty"` // 是否小端序
	Rate         float64 `json:"rate,omitempty"`          // 转换系数
	Correct      float64 `json:"correct,omitempty"`       // 校准值
	Bit          int     `json:"bit,omitempty"`           // 取位（>=1 取第 n 位）
	Enums        []Enum  `json:"enums,omitempty"`         // 枚举值定义
}

// Enum 枚举值定义
type Enum struct {
	Key   string `json:"key,omitempty"`
	Name  string `json:"name,omitempty"`
	Value int    `json:"value,omitempty"`
}

// Poller 轮询器定义 —— 一次显式的协议级读请求
// 只负责生成查询指令，值解码参数（Rate/Correct/Bit/Format 等）走 Property→Mapper 链路
type Poller struct {
	Register string `json:"register,omitempty"`
	Address  string `json:"address,omitempty"`
	Length   int    `json:"length,omitempty"`
}

// EventParam 事件参数
type EventParam struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// Event 事件定义（设备触发的告警/上报）
type Event struct {
	Key    string       `json:"key,omitempty"`
	Name   string       `json:"name,omitempty"`
	Params []EventParam `json:"params,omitempty"`
}

// ServiceParam 服务参数
type ServiceParam struct {
	Key      string  `json:"key,omitempty"`
	Name     string  `json:"name,omitempty"`
	Type     string  `json:"type,omitempty"`
	Min      float64 `json:"min,omitempty"`
	Max      float64 `json:"max,omitempty"`
	Step     float64 `json:"step,omitempty"`
	Default  string  `json:"default,omitempty"`
	Register bool    `json:"register,omitempty"`
}

// Service 服务定义（设备可调用的功能）
type Service struct {
	Key    string         `json:"key,omitempty"`
	Name   string         `json:"name,omitempty"`
	Type   string         `json:"type,omitempty"` // switch / slider / button / form
	Params []ServiceParam `json:"params,omitempty"`
}

// SettingParam 配置参数
type SettingParam struct {
	Key      string  `json:"key,omitempty"`
	Name     string  `json:"name,omitempty"`
	Type     string  `json:"type,omitempty"`
	Min      float64 `json:"min,omitempty"`
	Max      float64 `json:"max,omitempty"`
	Step     float64 `json:"step,omitempty"`
	Default  string  `json:"default,omitempty"`
	Register bool    `json:"register,omitempty"`
}

// Setting 配置项定义
type Setting struct {
	Key    string         `json:"key,omitempty"`
	Name   string         `json:"name,omitempty"`
	Params []SettingParam `json:"params,omitempty"`
}

// ControlOption 控件选项
type ControlOption struct {
	Label string `json:"label,omitempty"`
	Value string `json:"value,omitempty"`
	Icon  string `json:"icon,omitempty"`
}

// ControlEventHandler 控件事件处理器
type ControlEventHandler struct {
	Type   string            `json:"type,omitempty"` // property / service
	Key    string            `json:"key,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// Control 控件定义（前端 UI 控件）
type Control struct {
	ID      string                         `json:"id,omitempty"`
	Type    string                         `json:"type,omitempty"` // button / switch / slider / brightness / gauge / number / ...
	Name    string                         `json:"name,omitempty"`
	Span    int                            `json:"span,omitempty"`
	Confirm bool                           `json:"confirm,omitempty"`
	Binds   map[string]string              `json:"binds,omitempty"`
	Events  map[string]ControlEventHandler `json:"events,omitempty"`
	Styles  map[string]map[string]string   `json:"styles,omitempty"`
	Options []ControlOption                `json:"options,omitempty"`
	Min     float64                        `json:"min,omitempty"`
	Max     float64                        `json:"max,omitempty"`
	Step    float64                        `json:"step,omitempty"`
	Unit    string                         `json:"unit,omitempty"`
}

// Expression 表达式定义
type Expression struct {
	Key     string `json:"key,omitempty"`
	Name    string `json:"name,omitempty"`
	Expr    string `json:"expr,omitempty"`
	program *vm.Program
	err     error // 编译错误
}

// Filter 过滤器定义（计算结果应该是bool，为true时通过过滤）
type Filter struct {
	Name    string `json:"name,omitempty"`
	Expr    string `json:"expr,omitempty"`
	program *vm.Program
	err     error // 编译错误
}

// Model 通用物模型定义（与 pkg/protocol/model.go 的 RegisterModel 配合使用）
type Model struct {
	Protocol    string       `json:"protocol,omitempty" xorm:"protocol varchar(64)"`
	Properties  []Property   `json:"properties,omitempty" xorm:"properties json"`
	Pollers     []Poller     `json:"pollers,omitempty" xorm:"pollers json"`
	Expressions []Expression `json:"expressions,omitempty" xorm:"expressions json"`
	Filters     []Filter     `json:"filters,omitempty" xorm:"filters json"`
	Events      []Event      `json:"events,omitempty" xorm:"events json"`
	Services    []Service    `json:"services,omitempty" xorm:"services json"`
	Settings    []Setting    `json:"settings,omitempty" xorm:"settings json"`
	Controls    []Control    `json:"controls,omitempty" xorm:"controls json"`
}
