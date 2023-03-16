package model

type Group struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Desc    string `json:"desc,omitempty"`
	Created Time   `json:"created" xorm:"created"`
}