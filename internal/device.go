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
	"github.com/god-jason/iot-master/protocol"
	"github.com/god-jason/iot-master/space"
	"math/rand"
	"strconv"
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

	//waitingResponse map[string]chan any
	//waitingLock     sync.RWMutex
	waiting lib.Map[chan any]
}

func (d *Device) Open() error {

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
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/sync", &req)
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
	}
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/read", &req)
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
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/write", &req)
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
	token := mqtt.Publish("protocol/"+d.protocol+"/"+d.linker+"/"+d.LinkId+"/action", &req)
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
