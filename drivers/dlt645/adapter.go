package dlt645

import (
	"errors"
	"fmt"
	"github.com/zgwit/iot-master/log"
	"github.com/zgwit/iot-master/mqtt"
	"github.com/zgwit/iot-master/pool"
	"github.com/zgwit/iot-master/product"
	"github.com/zgwit/iot-master/types"
	"time"
)

type Adapter struct {
	dlt645 Dlt645

	devices  map[string]string
	stations map[string]types.Options

	mapper *Mapper
}

func (adapter *Adapter) Mount(deviceId string, productId string, station types.Options) (err error) {
	adapter.devices[deviceId] = productId
	adapter.stations[deviceId] = station

	//加载映射表
	adapter.mappers[productId], err = product.LoadConfig[Mapper](productId, "mapper")
	if err != nil {
		return err
	}

	//加载轮询表
	adapter.pollers[productId], err = product.LoadConfig[[]*Poller](productId, "poller")
	if err != nil {
		return err
	}

	return nil
}

func (adapter *Adapter) Unmount(deviceId string) error {
	delete(adapter.devices, deviceId)
	delete(adapter.stations, deviceId)
	return nil
}

func (adapter *Adapter) poll() {

	//设备上线
	//!!! 不能这样做，不然启动服务器会产生大量的消息
	//for _, dev := range adapter.index {
	//	topic := fmt.Sprintf("device/online/%s", dev.Slave)
	//	_ = mqtt.Publish(topic, nil)
	//}

	interval := adapter.options.Int64("poller_interval", 60) //默认1分钟轮询一次
	if interval < 1 {
		interval = 1
	}

	//按毫秒计时
	interval *= 1000

	//OUT:
	for {
		start := time.Now().UnixMilli()
		for _, dev := range adapter.devices {
			values, err := adapter.Sync(dev.Id)
			if err != nil {
				log.Error(err)
				continue
			}

			//d := device.Get(dev.Slave)
			if values != nil && len(values) > 0 {
				_ = pool.Insert(func() {
					topic := fmt.Sprintf("device/"+dev.Id+"/values", dev.Id)
					mqtt.Publish(topic, values)
				})
			}
		}

		//检查连接，避免空等待
		if !adapter.tunnel.Available() {
			break
		}

		//轮询间隔
		now := time.Now().UnixMilli()
		elapsed := now - start
		if elapsed < interval {
			time.Sleep(time.Millisecond * time.Duration(interval-elapsed))
		}

		//避免空转，睡眠1分钟（延迟10ms太长，睡1分钟也有点长）
		if elapsed < 10 {
			time.Sleep(time.Minute)
		}
	}

	log.Info("modbus adapter quit", adapter.tunnel.ID())

	//设备下线
	//for _, dev := range adapter.devices {
	//	topic := fmt.Sprintf("device/%s/offline", dev.Slave)
	//	_ = mqtt.Publish(topic, nil)
	//}

	//TODO d.SetAdapter(nil)
}

func (adapter *Adapter) Get(id, name string) (any, error) {
	d := adapter.index[id]
	station := d.Station.Slave

	addr, err := d.mapper.Lookup(name)
	if err != nil {
		return nil, err
	}

	return adapter.dlt645.Read(station, addr)
}

func (adapter *Adapter) Set(id, name string, value any) error {
	return errors.New("不支持写入")
}

func (adapter *Adapter) Sync(id string) (map[string]any, error) {
	d := adapter.index[id]
	station := d.Station.Slave

	//没有地址表，则跳过
	if d.mapper == nil {
		return nil, nil
	}

	values := make(map[string]any)
	for _, point := range d.mapper.Points {
		data, err := adapter.dlt645.Read(station, point.Id)
		if err != nil {
			return nil, err
		}
		values[point.Name] = data
	}
	//TODO 计算器

	return values, nil
}
