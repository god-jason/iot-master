package modbus

import (
	"errors"
	"github.com/god-jason/iot-master/product"
	"github.com/god-jason/iot-master/types"
)

type Adapter struct {
	modbus Modbus

	//device=>product_id
	devices  map[string]string
	stations map[string]types.Options

	//product_id => xxx
	mappers map[string]*Mapper
	pollers map[string]*[]*Poller
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

func (adapter *Adapter) Get(deviceId, name string) (any, error) {
	productId, has := adapter.devices[deviceId]
	if !has {
		return nil, errors.New("设备未注册")
	}
	station := adapter.stations[deviceId]
	slave := station.Int("slave", 1)

	mapper := adapter.mappers[productId]
	if mapper == nil {
		return nil, errors.New("没有地址映射")
	}
	point, code, address := mapper.Lookup(name)
	if point == nil {
		return nil, errors.New("找不到数据点")
	}

	//此处全部读取了，有些冗余
	data, err := adapter.modbus.Read(uint8(slave), code, address, 2)
	if err != nil {
		return nil, err
	}

	return point.Parse(address, data)
}

func (adapter *Adapter) Set(deviceId, name string, value any) error {
	productId, has := adapter.devices[deviceId]
	if !has {
		return errors.New("设备未注册")
	}
	station := adapter.stations[deviceId]
	slave := station.Int("slave", 1)

	mapper := adapter.mappers[productId]
	if mapper == nil {
		return errors.New("没有地址映射")
	}
	point, code, address := mapper.Lookup(name)
	if point == nil {
		return errors.New("地址找不到")
	}

	data, err := point.Encode(value)
	if err != nil {
		return err
	}
	return adapter.modbus.Write(uint8(slave), code, address, data)
}

func (adapter *Adapter) Poll(deviceId string) (map[string]any, error) {
	productId, has := adapter.devices[deviceId]
	if !has {
		return nil, errors.New("设备未注册")
	}
	station := adapter.stations[deviceId]
	slave := station.Int("slave", 1)

	//没有地址表和轮询器，则跳过
	//if d.pollers == nil || d.mappers == nil {
	//	return nil, nil
	//}
	mapper := adapter.mappers[productId]
	if mapper == nil {
		return nil, errors.New("没有地址映射")
	}

	values := make(map[string]any)
	for _, poller := range *adapter.pollers[productId] {
		if poller == nil {
			continue
		}
		data, err := adapter.modbus.Read(uint8(slave), poller.Code, poller.Address, poller.Length)
		if err != nil {
			return nil, err
		}
		err = poller.Parse(mapper, data, values)
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}
