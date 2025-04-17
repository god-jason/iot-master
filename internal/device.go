package internal

import (
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/god-jason/iot-master/device"
	"github.com/god-jason/iot-master/product"
	"github.com/god-jason/iot-master/project"
	"github.com/god-jason/iot-master/space"
	"sync"
	"time"
)

var devices lib.Map[Device]

func GetDevice(id string) *Device {
	return devices.Load(id)
}

type Device struct {
	device.Device `xorm:"extends"`

	valuesLock sync.RWMutex
	Values     map[string]any `json:"values"`
	Updated    time.Time      `json:"updated"`

	projects []string
	spaces   []string

	validators []*Validator
}

func (d *Device) Open() error {
	d.Values = make(map[string]any)

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
		err = vv.Build() //重复编译了
		if err != nil {
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
	topics = append(topics, "device/"+d.Id+"/values")
	for _, p := range d.projects {
		topics = append(topics, "project/"+p+"/device/"+d.Id+"/property")
	}
	for _, s := range d.spaces {
		topics = append(topics, "space/"+s+"/device/"+d.Id+"/property")
	}
	if len(topics) > 0 {
		mqtt.PublishEx(topics, values)
	}

	d.valuesLock.Lock()
	defer d.valuesLock.Unlock() //TODO 后续发消息和入库，锁的时间比较长

	//更新数据
	for k, v := range values {
		d.Values[k] = v
	}
	d.Values["__update"] = time.Now()

	//检查属性
	for _, v := range d.validators {
		alarm, err := v.Evaluate(d.Values)
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
