package iot

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/lib"
	"github.com/god-jason/iot-master/pkg/log"
	"github.com/god-jason/iot-master/pkg/mqtt"
)

type Property struct {
	Time  int64 `json:"time,omitempty"`
	Value any   `json:"value,omitempty"`
}

type Device struct {
	//device.Device `xorm:"extends"`
	Id        string `json:"id,omitempty" xorm:"pk"`
	GatewayId string `json:"gateway_id,omitempty" xorm:"index"`
	ProductId string `json:"product_id,omitempty" xorm:"index"`
	GroupId   string `json:"group_id,omitempty" xorm:"index"`
	LinkId    string `json:"link_id,omitempty" xorm:"index"`
	Name      string `json:"name,omitempty"`
	Disabled  bool   `json:"disabled,omitempty"` //禁用
	Online    bool   `json:"online,omitempty"`
	Error     string `json:"error,omitempty"`

	values Values

	linker   string
	protocol string

	validators []*Validator

	//waitingResponse map[string]chan any
	//waitingLock     sync.RWMutex
	waiting lib.Map[chan any]
}

type Status struct {
	Online bool   `json:"online,omitempty"`
	Error  string `json:"error,omitempty"`
}

func (d *Device) Open() error {
	//d.Online = true

	//加载产品物模型
	productModel, err := LoadModel(d.ProductId)
	if err != nil {
		return err
	}

	//复制
	for _, v := range productModel.Validators {
		if v.Disabled {
			continue
		}
		vv := &Validator{Validator: v}
		d.validators = append(d.validators, vv)
		err = vv.Build() //TODO 重复编译了
		if err != nil {
			d.Error = err.Error()
			log.Error(err)
		}
	}

	return nil
}

func (d *Device) PutValues(values map[string]any) {

	//TODO 过滤器实现

	//广播消息
	var topics []string

	//发给历史数据库
	topics = append(topics, "history/"+d.ProductId+"/"+d.Id+"/values")

	mqtt.PublishEx(topics, values)

	d.values.Put(values)

	//检查属性
	for _, v := range d.validators {
		alarm, err := v.Evaluate(d.values.Get())
		if err != nil {
			log.Error(err)
		}
		if alarm != nil {
			alarm.DeviceId = d.Id

			var topics []string
			topics = append(topics, "device/"+d.Id+"/alarm")

			//入数据库
			_, err = db.Engine().InsertOne(alarm)
			if err != nil {
				log.Error(err)
			}

			mqtt.PublishEx(topics, alarm)
		}
	}
}

func (d *Device) GetValues() map[string]any {
	return d.values.Get()
}

func (d *Device) waitResponse(msg_id string, timeout int) (any, error) {
	//等待消息
	ch := make(chan any)

	c := d.waiting.LoadAndStore(msg_id, &ch)
	if c != nil {
		close(*c)
	}

	if timeout < 1 {
		timeout = 30
	}

	select {
	case resp := <-ch:
		d.waiting.Delete(msg_id)
		return resp, nil
	case <-time.After(time.Duration(timeout) * time.Second):
		d.waiting.Delete(msg_id)
		return nil, errors.New("请求超时")
	}
}

func (d *Device) Sync(timeout int) (map[string]any, error) {
	req := SyncRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
	}

	mqtt.Publish("device/"+d.Id+"/sync", req)
	if d.GatewayId != "" {
		mqtt.Publish("device/"+d.GatewayId+"/sync", req)
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*SyncResponse); ok {
		if res.Error != "" {
			return nil, errors.New(res.Error)
		}
		return res.Values, nil
	} else {
		return nil, errors.New("want type SyncResponse")
	}
}

func (d *Device) onSyncResponse(resp *SyncResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}

func (d *Device) Read(points []string, timeout int) (map[string]any, error) {
	req := ReadRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
		Points:   points,
	}
	mqtt.Publish("device/"+d.Id+"/read", req)
	if d.GatewayId != "" {
		mqtt.Publish("device/"+d.GatewayId+"/sub/"+d.Id+"/read", req)
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*ReadResponse); ok {
		if res.Error != "" {
			return nil, errors.New(res.Error)
		}
		return res.Values, nil
	} else {
		return nil, errors.New("want type ReadResponse")
	}
}

func (d *Device) onReadResponse(resp *ReadResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}

func (d *Device) Write(values map[string]any, timeout int) (map[string]bool, error) {
	req := WriteRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
		Values:   values,
	}
	mqtt.Publish("device/"+d.Id+"/write", req)
	if d.GatewayId != "" {
		mqtt.Publish("device/"+d.GatewayId+"/sub/"+d.Id+"/write", req)
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*WriteResponse); ok {
		if res.Error != "" {
			return nil, errors.New(res.Error)
		}
		return res.Result, nil
	} else {
		return nil, errors.New("want type WriteResponse")
	}
}

func (d *Device) onWriteResponse(resp *WriteResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}

func (d *Device) Action(action string, parameters map[string]any, timeout int) (map[string]any, error) {
	req := ActionRequest{
		MsgId:      strconv.FormatInt(rand.Int63(), 10),
		DeviceId:   d.Id,
		Action:     action,
		Parameters: parameters,
	}

	//兼容旧设备，TODO 后续需要删除
	mqtt.Publish("device/"+d.Id+"/action/"+action, parameters)

	//发送消息
	mqtt.Publish("device/"+d.Id+"/action", req)
	if d.GatewayId != "" {
		mqtt.Publish("device/"+d.GatewayId+"/sub/"+d.Id+"/action", req)
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*ActionResponse); ok {
		if res.Error != "" {
			return nil, errors.New(res.Error)
		}
		return res.Result, nil
	} else {
		return nil, errors.New("want type ActionResponse")
	}
}

func (d *Device) onActionResponse(resp *ActionResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}
