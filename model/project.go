package model

import (
	"github.com/zgwit/iot-master/calc"
	"time"
)

//Project 项目
type Project struct {
	ID       int    `json:"id" storm:"id,increment"`
	Disabled bool   `json:"disabled,omitempty"`
	Template string `json:"template,omitempty"`

	Devices []*ProjectDevice `json:"devices"`
	//Devices []int `json:"devices"`

	Aggregators []*Aggregator `json:"aggregators"`
	Commands    []*Command    `json:"commands"`
	Jobs        []*Job        `json:"jobs"`
	Strategies  []*Strategy   `json:"strategies"`

	Context calc.Context `json:"context"`
}

//ProjectDevice 项目的设备
type ProjectDevice struct {
	ID   int    `json:"id"`
	Name string `json:"name"` //编程名
}

//ProjectHistory 项目历史
type ProjectHistory struct {
	ID        int       `json:"id" storm:"id,increment"`
	ProjectID int       `json:"project_id"`
	History   string    `json:"history"`
	Created   time.Time `json:"created"`
}

//ProjectHistoryAlarm 项目历史告警
type ProjectHistoryAlarm struct {
	ID int `json:"id" storm:"id,increment"`

	ProjectID int    `json:"project_id"`
	DeviceID  int    `json:"device_id"`
	Code      string `json:"code"`
	Level     int    `json:"level"`
	Message   string `json:"message"`

	Created time.Time `json:"created"`
}

//ProjectHistoryReactor 项目历史响应
type ProjectHistoryReactor struct {
	ID        int       `json:"id" storm:"id,increment"`
	ProjectID int       `json:"project_id"`
	Name      string    `json:"name"`
	History   string    `json:"result"`
	Created   time.Time `json:"created"`
}

//ProjectHistoryJob 项目历史任务
type ProjectHistoryJob struct {
	ID        int       `json:"id" storm:"id,increment"`
	ProjectID int       `json:"project_id"`
	Job       string    `json:"job"`
	History   string    `json:"result"`
	Created   time.Time `json:"created"`
}