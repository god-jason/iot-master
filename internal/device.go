package internal

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/product"
	"github.com/god-jason/iot-master/protocol"
)

type Property struct {
	Time  int64 `json:"time,omitempty"`
	Value any   `json:"value,omitempty"`
}

type Device struct {
	//device.Device `xorm:"extends"`
	Id        string         `json:"id,omitempty" xorm:"pk"`
	ProductId string         `json:"product_id,omitempty" xorm:"index"`
	LinkId    string         `json:"link_id,omitempty" xorm:"index"`
	Name      string         `json:"name,omitempty"`
	Station   map[string]any `json:"station,omitempty" xorm:"json"` //从站信息（协议定义表单）
	Disabled  bool           `json:"disabled,omitempty"`            //禁用

	Status `xorm:"-"`

	values Values

	linker   string
	protocol string

	projects []string
	spaces   []string

	validators []*Validator

	//waitingResponse map[string]chan any
	//waitingLock     sync.RWMutex
	waiting lib.Map[chan any]
}

type DeviceModel struct {
	Id         string               `json:"id,omitempty" xorm:"pk"`
	Validators []*product.Validator `json:"validators,omitempty" xorm:"json"`
	Created    time.Time            `json:"created,omitempty" xorm:"created"`
}

type Status struct {
	Online bool   `json:"online,omitempty"`
	Error  string `json:"error,omitempty"`
}

func (d *Device) Open() error {
	d.Online = true

	//查询绑定的项目
	var ps []map[string]interface{}
	err := db.Engine().Table("project_device").Cols("project_id").Where("device_id=?", d.Id).Find(&ps) //.Distinct("project_id")
	if err != nil {
		return err
	}
	for _, p := range ps {
		d.projects = append(d.projects, p["project_id"].(string))
	}

	//查询绑定的设备
	var ss []map[string]interface{}
	err = db.Engine().Table("space_device").Cols("space_id").Where("device_id=?", d.Id).Find(&ss) //.Distinct("space_id")
	if err != nil {
		return err
	}
	for _, s := range ss {
		d.spaces = append(d.spaces, s["space_id"].(string))
	}

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

	//加载设备模型
	var deviceModel DeviceModel
	has, err := db.Engine().ID(d.Id).Get(&deviceModel)
	if err != nil {
		return err
	}
	if has {
		for _, v := range deviceModel.Validators {
			if v.Disabled {
				continue
			}
			vv := &Validator{Validator: v}
			d.validators = append(d.validators, vv)
			err = vv.Build() //重复编译了
			if err != nil {
				d.Error = err.Error()
				log.Error(err)
			}
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

	//发给项目
	for _, p := range d.projects {
		topics = append(topics, "project/"+p+"/device/"+d.Id+"/values")
	}
	//发给空间
	for _, s := range d.spaces {
		topics = append(topics, "space/"+s+"/device/"+d.Id+"/values")
	}

	mqtt.PublishEx(topics, values)

	d.values.Put(values)

	//检查属性
	for _, v := range d.validators {
		alarm, err := v.Evaluate(d.values.Get())
		if err != nil {
			log.Error(err)
		}
		if alarm != nil {
			alarm.Device = d.Name
			alarm.DeviceId = d.Id

			var topics []string
			topics = append(topics, "device/"+d.Id+"/alarm")
			for _, p := range d.projects {
				alarm.ProjectId = p //TODO 多项目，会被覆盖掉
				topics = append(topics, "project/"+p+"/device/"+d.Id+"/alarm")
			}

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
	req := protocol.SyncRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/link/"+d.linker+"/"+d.LinkId+"/sync", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*protocol.SyncResponse); ok {
		return res.Values, nil
	} else {
		return nil, errors.New("want type SyncResponse")
	}
}

func (d *Device) onSyncResponse(resp *protocol.SyncResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}

func (d *Device) Read(points []string, timeout int) (map[string]any, error) {
	req := protocol.ReadRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
		Points:   points,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/link/"+d.linker+"/"+d.LinkId+"/read", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*protocol.ReadResponse); ok {
		return res.Values, nil
	} else {
		return nil, errors.New("want type ReadResponse")
	}
}

func (d *Device) onReadResponse(resp *protocol.ReadResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}

func (d *Device) Write(values map[string]any, timeout int) (map[string]bool, error) {
	req := protocol.WriteRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
		Values:   values,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/link/"+d.linker+"/"+d.LinkId+"/write", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*protocol.WriteResponse); ok {
		return res.Result, nil
	} else {
		return nil, errors.New("want type WriteResponse")
	}
}

func (d *Device) onWriteResponse(resp *protocol.WriteResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}

func (d *Device) Action(action string, parameters map[string]any, timeout int) (map[string]any, error) {
	req := protocol.ActionRequest{
		MsgId:      strconv.FormatInt(rand.Int63(), 10),
		DeviceId:   d.Id,
		Action:     action,
		Parameters: parameters,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/link/"+d.linker+"/"+d.LinkId+"/action", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	resp, err := d.waitResponse(req.MsgId, timeout)
	if err != nil {
		return nil, err
	}

	if res, ok := resp.(*protocol.ActionResponse); ok {
		return res.Result, nil
	} else {
		return nil, errors.New("want type ActionResponse")
	}
}

func (d *Device) onActionResponse(resp *protocol.ActionResponse) {
	c := d.waiting.LoadAndDelete(resp.MsgId)
	if c != nil {
		*c <- resp
	}
}
