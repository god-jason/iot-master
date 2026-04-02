package iot

import "time"

type Alarm struct {
	Id       int64     `json:"id,omitempty"`
	DeviceId string    `json:"device_id,omitempty" xorm:"index"`
	GroupId  string    `json:"group_id,omitempty" xorm:"index"`
	Title    string    `json:"title,omitempty"`
	Message  string    `json:"message,omitempty"`
	Level    int       `json:"level,omitempty"`
	Created  time.Time `json:"created,omitempty" xorm:"created"`
}
