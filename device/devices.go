package device

import (
	"errors"
	"github.com/god-jason/iot-master/db"
	"github.com/god-jason/iot-master/lib"
	"github.com/god-jason/iot-master/log"
)

func LoadByTunnel2[T any](id string) (devices []T, err error) {
	err = db.Engine.Where("tunnel_id=?", id).And("disabled!=1").Find(&devices)
	return
}

var devices lib.Map[Device]

func Load(device *Device) error {
	dev := devices.LoadAndStore(device.Id, device)
	if dev != nil {
		//dev.Destroy()
	}
	return device.Open()
}

func LoadByTunnel(id string) (devices []*Device, err error) {
	//devices []*Device
	err = db.Engine.Where("tunnel_id=?", id).And("disabled!=1").Find(&devices)
	for _, device := range devices {
		err := Load(device)
		if err != nil {
			log.Error(err)
			//return nil, err
		}
	}

	return
}

func Ensure(id string) (*Device, error) {
	var device Device
	has, err := db.Engine.ID(id).Get(&device)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("找不到设备")
	}
	err = Load(&device)
	if err != nil {
		return nil, err
	}
	return &device, nil
}
