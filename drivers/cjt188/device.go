package cjt188

import "github.com/god-jason/iot-master/db"

func init() {
	db.Register(new(Device))
}

type Station struct {
	Slave string `json:"slave"`
}

type Device struct {
	Id        string `json:"id" xorm:"pk"`
	ProductId string `json:"product_id"`

	//cjt188站号
	Station Station `json:"station,omitempty" xorm:"json"`

	//映射和轮询表
	mapper *Mapper
}
