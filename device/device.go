package device

import (
	"errors"
	"github.com/god-jason/bucket/pkg/exception"
	"github.com/god-jason/bucket/types"
	"github.com/god-jason/iot-master/base"
	"github.com/god-jason/iot-master/db"
	"github.com/god-jason/iot-master/driver"
	"github.com/god-jason/iot-master/mqtt"
	"github.com/god-jason/iot-master/product"
	"time"
)

func init() {
	db.Register(new(Device))
}

type Device struct {
	Id string `json:"id" xorm:"pk"` //ClientID

	ProductId string `json:"product_id,omitempty" xorm:"index"`
	Product   string `json:"product,omitempty" xorm:"<-"`

	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty" xorm:"json"` //模型参数，用于报警检查
	Disabled    bool           `json:"disabled,omitempty"`
	Created     time.Time      `json:"created,omitempty" xorm:"created"`

	Online bool `json:"online,omitempty" xorm:"-"`

	//通道ID
	TunnelId string        `json:"tunnel_id,omitempty" xorm:"index"`
	Station  types.Options `json:"station,omitempty" xorm:"json"` //通道参数 保存从站号等

	//变量
	values map[string]any
	//last   time.Time

	adapter driver.Adapter

	actions map[string]*base.Action
}

func (d *Device) Open() error {
	operators, err := product.LoadConfig[[]*base.Action](d.ProductId, "actions")
	if err != nil {
		return err
	}

	d.actions = make(map[string]*base.Action)
	for _, op := range *operators {
		d.actions[op.Name] = op
		err = op.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) Values() map[string]any {
	return d.values
}

func (d *Device) Push(values map[string]any) {

	//赋值
	if d.values == nil {
		d.values = make(map[string]any)
	}
	for k, v := range values {
		d.values[k] = v
	}

	topic := "up/device/" + d.Id + "/property"
	mqtt.Publish(topic, values)
	//todo 上传失败，保存历史
}

func (d *Device) Write(point string, value any) error {
	if d.adapter == nil {
		return errors.New("未连接")
	}
	return d.adapter.Set(d.Id, point, value)
}

func (d *Device) WriteMany(values map[string]any) error {
	if d.adapter == nil {
		return errors.New("未连接")
	}
	for point, value := range values {
		err := d.adapter.Set(d.Id, point, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) SetAdapter(adapter driver.Adapter) {
	d.adapter = adapter //TODO 会内存泄露，需要手动清空
}

func (d *Device) Action(name string, values map[string]any) (map[string]any, error) {
	action := d.actions[name]
	if action == nil {
		return nil, exception.New("找不到操作")
	}

	executors, err := action.GetExecutors(values)
	if err != nil {
		return nil, err
	}

	for _, executor := range executors {
		if executor.Delay > 0 {
			time.AfterFunc(executor.Delay*time.Millisecond, func() {
				_ = d.Write(executor.Point, executor.Value)
			})
		} else {
			err = d.Write(executor.Point, executor.Value)
			if err != nil {
				return nil, err
			}
		}
	}

	//todo 取返回值

	return nil, nil
}
