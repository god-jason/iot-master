package internal

import (
	"errors"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/device"
	"github.com/god-jason/iot-master/link"
	"github.com/god-jason/iot-master/product"
	"github.com/god-jason/iot-master/project"
	"github.com/god-jason/iot-master/space"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var devices lib.Map[Device]

func GetDevice(id string) *Device {
	return devices.Load(id)
}

type Device struct {
	device.Device
	device.Status

	values Values

	linker   string
	protocol string

	projects []string
	spaces   []string

	validators []*Validator

	waitingResponse map[string]chan any
	waitingLock     sync.RWMutex
}

func (d *Device) Open() error {
	d.waitingResponse = make(map[string]chan any)

	//加载连接(主要是协议)
	if d.LinkId != "" {
		var lnk link.Link
		has, err := db.Engine().ID(d.LinkId).Get(&lnk)
		if err != nil {
			d.Error = err.Error()
			return err
		}
		if !has {
			d.Error = "没有指定链接"
			return errors.New(d.Error)
		}
		d.protocol = lnk.Protocol
		d.linker = lnk.Linker
	}

	//查询绑定的项目
	var ps []*project.ProjectDevice
	err := db.Engine().Where("device_id=?", d.Id).Find(&ps) //.Distinct("project_id")
	if err != nil {
		return err
	}
	for _, p := range ps {
		d.projects = append(d.projects, p.ProjectId)
	}

	//查询绑定的设备
	var ss []*space.SpaceDevice
	err = db.Engine().Where("device_id=?", d.Id).Find(&ss) //.Distinct("space_id")
	if err != nil {
		return err
	}
	for _, s := range ss {
		d.spaces = append(d.spaces, s.SpaceId)
	}

	//加载产品物模型
	productModel, err := product.LoadModel(d.ProductId)
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
	var deviceModel device.DeviceModel
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
	for _, p := range d.projects {
		topics = append(topics, "project/"+p+"/device/"+d.Id+"/values")
	}
	for _, s := range d.spaces {
		topics = append(topics, "space/"+s+"/device/"+d.Id+"/values")
	}
	if len(topics) > 0 {
		mqtt.PublishEx(topics, values)
	}

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

type SyncRequest struct {
	MsgId    string `json:"msg_id"`
	DeviceId string `json:"device_id"`
}

type SyncResponse struct {
	MsgId    string         `json:"msg_id"`
	DeviceId string         `json:"device_id"`
	Values   map[string]any `json:"values"`
}

func (d *Device) Sync(timeout int) (map[string]any, error) {
	req := SyncRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/sync", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	//等待消息
	ch := make(chan any)

	d.waitingLock.Lock()
	d.waitingResponse[req.MsgId] = ch
	d.waitingLock.Unlock()

	if timeout < 1 {
		timeout = 30
	}

	select {
	case resp := <-ch:
		if res, ok := resp.(*SyncResponse); ok {
			return res.Values, nil
		} else {
			return nil, errors.New("want type SyncResponse")
		}
	case <-time.After(time.Duration(timeout) * time.Second):

		d.waitingLock.Lock()
		delete(d.waitingResponse, req.MsgId)
		d.waitingLock.Unlock()

		return nil, errors.New("请求超时")
	}
}

func (d *Device) onSyncResponse(resp *SyncResponse) {
	d.waitingLock.RLock()
	defer d.waitingLock.RUnlock()

	if ch, ok := d.waitingResponse[resp.MsgId]; ok {
		ch <- ch
	}
}

type ReadRequest struct {
	MsgId    string   `json:"msg_id"`
	DeviceId string   `json:"device_id"`
	Points   []string `json:"points"`
}

type ReadResponse struct {
	MsgId    string         `json:"msg_id"`
	DeviceId string         `json:"device_id"`
	Values   map[string]any `json:"values"`
}

func (d *Device) Read(points []string, timeout int) (map[string]any, error) {
	req := ReadRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/read", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	//等待消息
	ch := make(chan any)

	d.waitingLock.Lock()
	d.waitingResponse[req.MsgId] = ch
	d.waitingLock.Unlock()

	if timeout < 1 {
		timeout = 30
	}

	select {
	case resp := <-ch:
		if res, ok := resp.(*ReadResponse); ok {
			return res.Values, nil
		} else {
			return nil, errors.New("want type ReadResponse")
		}
	case <-time.After(time.Duration(timeout) * time.Second):

		d.waitingLock.Lock()
		delete(d.waitingResponse, req.MsgId)
		d.waitingLock.Unlock()

		return nil, errors.New("请求超时")
	}
}

func (d *Device) onReadResponse(resp *ReadResponse) {
	d.waitingLock.RLock()
	defer d.waitingLock.RUnlock()

	if ch, ok := d.waitingResponse[resp.MsgId]; ok {
		ch <- ch
	}
}

type WriteRequest struct {
	MsgId    string         `json:"msg_id"`
	DeviceId string         `json:"device_id"`
	Values   map[string]any `json:"values"`
}

type WriteResponse struct {
	MsgId    string          `json:"msg_id"`
	DeviceId string          `json:"device_id"`
	Result   map[string]bool `json:"result"`
}

func (d *Device) Write(values map[string]any, timeout int) (map[string]bool, error) {
	req := WriteRequest{
		MsgId:    strconv.FormatInt(rand.Int63(), 10),
		DeviceId: d.Id,
		Values:   values,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/write", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	//等待消息
	ch := make(chan any)

	d.waitingLock.Lock()
	d.waitingResponse[req.MsgId] = ch
	d.waitingLock.Unlock()

	if timeout < 1 {
		timeout = 30
	}

	select {
	case resp := <-ch:
		if res, ok := resp.(*WriteResponse); ok {
			return res.Result, nil
		} else {
			return nil, errors.New("want type WriteResponse")
		}
	case <-time.After(time.Duration(timeout) * time.Second):

		d.waitingLock.Lock()
		delete(d.waitingResponse, req.MsgId)
		d.waitingLock.Unlock()

		return nil, errors.New("请求超时")
	}
}

func (d *Device) onWriteResponse(resp *WriteResponse) {
	d.waitingLock.RLock()
	defer d.waitingLock.RUnlock()

	if ch, ok := d.waitingResponse[resp.MsgId]; ok {
		ch <- ch
	}
}

type ActionRequest struct {
	MsgId      string         `json:"msg_id"`
	DeviceId   string         `json:"device_id"`
	Action     string         `json:"action"`
	Parameters map[string]any `json:"parameters"`
}

type ActionResponse struct {
	MsgId    string         `json:"msg_id"`
	DeviceId string         `json:"device_id"`
	Result   map[string]any `json:"result"`
}

func (d *Device) Action(action string, parameters map[string]any, timeout int) (map[string]any, error) {
	req := ActionRequest{
		MsgId:      strconv.FormatInt(rand.Int63(), 10),
		DeviceId:   d.Id,
		Action:     action,
		Parameters: parameters,
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/action", &req)
	token.Wait()
	err := token.Error()
	if err != nil {
		return nil, err
	}

	//等待消息
	ch := make(chan any)

	d.waitingLock.Lock()
	d.waitingResponse[req.MsgId] = ch
	d.waitingLock.Unlock()

	if timeout < 1 {
		timeout = 30
	}

	select {
	case resp := <-ch:
		if res, ok := resp.(*ActionResponse); ok {
			return res.Result, nil
		} else {
			return nil, errors.New("want type ActionResponse")
		}
	case <-time.After(time.Duration(timeout) * time.Second):

		d.waitingLock.Lock()
		delete(d.waitingResponse, req.MsgId)
		d.waitingLock.Unlock()

		return nil, errors.New("请求超时")
	}
}

func (d *Device) onActionResponse(resp *ActionResponse) {
	d.waitingLock.RLock()
	defer d.waitingLock.RUnlock()

	if ch, ok := d.waitingResponse[resp.MsgId]; ok {
		ch <- ch
	}
}
