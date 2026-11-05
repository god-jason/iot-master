package mqtt

import (
	"errors"
	"fmt"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"

	"github.com/google/uuid"
)

// ErrNotSupported 此传输不支持的操作
var ErrNotSupported = errors.New("mqtt transport does not support this operation")

// 出站指令等待应答的超时时间
const (
	// defaultTimeout 属性读写、配置查询应答超时
	defaultTimeout = 10 * time.Second
	// controlTimeout 服务调用、配置下发应答超时
	controlTimeout = 30 * time.Second
)

// Device 外置 broker 客户端接入的 Device 实现。
// 组合 iot.BaseDevice（通用状态 + 入站），出站指令等待设备应答。
type Device struct {
	*iot.BaseDevice
}

// NewDevice 构造外置 broker Device
func NewDevice(dev *types.Device) *Device {
	return &Device{BaseDevice: iot.NewBaseDevice(dev)}
}

// 编译时断言 Device 完整实现了 iot.Device interface
var _ iot.Device = (*Device)(nil)

// ===== 出站指令（Publish 后等待设备 ACK，带超时）=====

// SetProperty 下发属性并等待设备应答
func (d *Device) SetProperty(props map[string]interface{}) error {
	id := uuid.New().String()
	resp, err := d.request(PropertySet, PropertySetMsg{ID: id, Params: props}, id, defaultTimeout)
	if err != nil {
		return err
	}
	return checkRet(resp)
}

// GetProperty 读取属性并等待设备应答
func (d *Device) GetProperty(names []string) (map[string]interface{}, error) {
	id := uuid.New().String()
	resp, err := d.request(PropertyGet, PropertyGetMsg{
		ID:     id,
		Params: map[string]interface{}{"properties": names},
	}, id, defaultTimeout)
	if err != nil {
		return nil, err
	}
	if err := checkRet(resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Control 调用设备服务并等待设备应答
func (d *Device) Control(service string, params map[string]interface{}) (map[string]interface{}, error) {
	id := uuid.New().String()
	resp, err := d.request(ServiceCall, ServiceCallMsg{ID: id, Service: service, Params: params}, id, controlTimeout)
	if err != nil {
		return nil, err
	}
	if err := checkRet(resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SetConfig 下发设备配置并等待设备应答
func (d *Device) SetConfig(data map[string]interface{}) error {
	id := uuid.New().String()
	resp, err := d.request(SettingSet, SettingSetMsg{ID: id, Setting: data}, id, controlTimeout)
	if err != nil {
		return err
	}
	return checkRet(resp)
}

// GetConfig 读取设备配置并等待设备应答
func (d *Device) GetConfig(keys []string) (map[string]interface{}, error) {
	id := uuid.New().String()
	resp, err := d.request(SettingGet, SettingGetMsg{ID: id, Keys: keys}, id, defaultTimeout)
	if err != nil {
		return nil, err
	}
	if err := checkRet(resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Poll 主动采集 — MQTT 接入是被动上报，不支持轮询
func (d *Device) Poll() (map[string]interface{}, error) {
	return nil, ErrNotSupported
}

// request 下发指令到 suffix 主题，按 id 等待设备应答；超时返回 ErrResponseTimeout
func (d *Device) request(suffix string, payload any, id string, timeout time.Duration) (*ResponseMessage, error) {
	ch, err := responseManager.Add(id)
	if err != nil {
		return nil, err
	}
	Publish(BuildTopic(d.ID(), suffix), payload)
	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		responseManager.Remove(id)
		return nil, ErrResponseTimeout
	}
}

// checkRet 校验设备应答的 ret 字段（1 成功 / 0 失败）
func checkRet(resp *ResponseMessage) error {
	if resp == nil {
		return errors.New("mqtt: empty response")
	}
	if resp.Ret != 1 {
		return fmt.Errorf("mqtt: device response failed: %s", resp.Message)
	}
	return nil
}
