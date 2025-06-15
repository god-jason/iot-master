package internal

import (
	"errors"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/lib"
)

var devices lib.Map[Device]

func GetDevice(id string) *Device {
	return devices.Load(id)
}

func LoadDevice(id string) (*Device, error) {
	d := &Device{}
	has, err := db.Engine().ID(id).Get(&d.Device)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("device not exist")
	}
	err = d.Open()
	if err != nil {
		return nil, err
	}
	devices.Store(id, d)

	return d, nil
}

func UnloadDevice(id string) error {
	//close?
	devices.Delete(id)
	return nil
}
