package broker

import (
	"errors"
	"fmt"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"

	"github.com/google/uuid"
)

// ErrNotSupported 此传输不支持的操作
var ErrNotSupported = errors.New("broker transport does not support this operation")

// Device 内嵌 broker 接入的 Device 实现。
// 组合 iot.BaseDevice（通用状态 + 入站），出站指令经 Service 下发并等待设备 ACK。
type Device struct {
	*iot.BaseDevice

	svc *Service // 所属消息服务，出站指令经其下发并等待应答
}

// NewDevice 构造内嵌 broker Device
func NewDevice(dev *types.Device, svc *Service) *Device {
	return &Device{
		BaseDevice: iot.NewBaseDevice(dev),
		svc:        svc,
	}
}

// 编译时断言 Device 完整实现了 iot.Device interface
var _ iot.Device = (*Device)(nil)

// ===== 出站指令（Service.Publish 后等待设备 ACK，带超时）=====

// SetProperty 下发属性并等待设备应答
func (d *Device) SetProperty(props map[string]interface{}) error {
	id := uuid.New().String()
	resp, err := d.svc.request(d.ID(), PropertySet, PropertySetMsg{ID: id, Params: props}, id, defaultTimeout)
	if err != nil {
		return err
	}
	return checkRet(resp)
}

// GetProperty 读取属性并等待设备应答
func (d *Device) GetProperty(names []string) (map[string]interface{}, error) {
	id := uuid.New().String()
	resp, err := d.svc.request(d.ID(), PropertyGet, PropertyGetMsg{
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
	resp, err := d.svc.request(d.ID(), ServiceCall, ServiceCallMsg{ID: id, Service: service, Params: params}, id, controlTimeout)
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
	resp, err := d.svc.request(d.ID(), SettingSet, SettingSetMsg{ID: id, Setting: data}, id, controlTimeout)
	if err != nil {
		return err
	}
	return checkRet(resp)
}

// GetConfig 读取设备配置并等待设备应答
func (d *Device) GetConfig(keys []string) (map[string]interface{}, error) {
	id := uuid.New().String()
	resp, err := d.svc.request(d.ID(), SettingGet, SettingGetMsg{ID: id, Keys: keys}, id, defaultTimeout)
	if err != nil {
		return nil, err
	}
	if err := checkRet(resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// SyncDatabase 向设备同步数据库表记录并等待应答
func (d *Device) SyncDatabase(table string, records []map[string]interface{}) (*ResponseMessage, error) {
	id := uuid.New().String()
	return d.svc.request(d.ID(), DatabaseSync, DatabaseSyncMsg{
		ID:     id,
		Params: DatabaseSyncParams{Table: table, Records: records},
	}, id, controlTimeout)
}

// LoadDatabase 读取设备本地数据库表并等待应答
func (d *Device) LoadDatabase(table string) (*ResponseMessage, error) {
	id := uuid.New().String()
	return d.svc.request(d.ID(), DatabaseLoad, DatabaseLoadMsg{
		ID:     id,
		Params: DatabaseLoadParams{Table: table},
	}, id, defaultTimeout)
}

// Poll 主动采集 — MQTT 接入是被动上报，不支持轮询
func (d *Device) Poll() (map[string]interface{}, error) {
	return nil, ErrNotSupported
}

// checkRet 校验设备应答的 ret 字段（1 成功 / 0 失败）
func checkRet(resp *ResponseMessage) error {
	if resp == nil {
		return errors.New("broker: empty response")
	}
	if resp.Ret != 1 {
		return fmt.Errorf("broker: device response failed: %s", resp.Message)
	}
	return nil
}
