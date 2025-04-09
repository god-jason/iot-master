package project

import (
	"github.com/busy-cloud/boat/db"
	"time"
)

func init() {
	db.Register(new(Project), new(ProjectUser), new(ProjectDevice), new(ProjectApp))
}

type Project struct {
	Id          string    `json:"id" xorm:"pk"`
	Name        string    `json:"name,omitempty"`        //名称
	Description string    `json:"description,omitempty"` //说明
	Keywords    []string  `json:"keywords,omitempty"`    //关键字
	Disabled    bool      `json:"disabled,omitempty"`
	Created     time.Time `json:"created" xorm:"created"`
}

type ProjectUser struct {
	ProjectId string    `json:"project_id,omitempty" xorm:"pk"`
	Project   string    `json:"project,omitempty" xorm:"<-"`
	UserId    string    `json:"user_id,omitempty" xorm:"pk"`
	User      string    `json:"user,omitempty" xorm:"<-"`
	Admin     bool      `json:"admin,omitempty"`
	Disabled  bool      `json:"disabled,omitempty"`
	Created   time.Time `json:"created" xorm:"created"`
}

type ProjectDevice struct {
	ProjectId string    `json:"project_id,omitempty" xorm:"pk"`
	Project   string    `json:"project,omitempty" xorm:"<-"`
	DeviceId  string    `json:"device_id,omitempty" xorm:"pk"`
	Device    string    `json:"device,omitempty" xorm:"<-"`
	Name      string    `json:"name,omitempty"` //编程别名
	Created   time.Time `json:"created" xorm:"created"`
}

type ProjectApp struct {
	ProjectId string    `json:"project_id,omitempty" xorm:"pk"`
	AppId     string    `json:"app_id,omitempty" xorm:"pk"`
	Disabled  bool      `json:"disabled,omitempty"`
	Created   time.Time `json:"created" xorm:"created"`
}
