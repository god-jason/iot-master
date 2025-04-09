package alarm

import (
	"github.com/busy-cloud/boat/db"
	"time"
)

func init() {
	db.Register(&Alarm{})
}

type Alarm struct {
	Id        int64     `json:"id,omitempty"`
	DeviceId  string    `json:"device_id,omitempty" xorm:"index"`
	ProjectId string    `json:"project_id,omitempty" xorm:"index"`
	Device    string    `json:"device,omitempty" xorm:"-"`
	Project   string    `json:"project,omitempty" xorm:"-"`
	Title     string    `json:"title,omitempty"`
	Message   string    `json:"message,omitempty"`
	Level     int       `json:"level,omitempty"`
	Created   time.Time `json:"created,omitempty" xorm:"created"`
}
