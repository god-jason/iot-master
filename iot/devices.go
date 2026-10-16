package iot

import (
	"errors"
	"sync"

	"github.com/god-jason/iot-master/pkg/db"
)

var devices sync.Map

func GetDevice(id string) *Device {
	if v, ok := devices.Load(id); ok {
		return v.(*Device)
	}
	return nil
}

func LoadDevice(id string) (*Device, error) {
	var d Device
	has, err := db.Engine().ID(id).Get(&d)
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
	devices.Store(id, &d)

	return &d, nil
}

func UnloadDevice(id string) error {
	//close?
	devices.Delete(id)
	return nil
}
