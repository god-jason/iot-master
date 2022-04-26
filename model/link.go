package model

import "time"

//Link 链接模型
type Link struct {
	Id       int       `json:"id" storm:"id,increment"`
	TunnelId int       `json:"tunnel_id" storm:"index"`
	SN       string    `json:"sn"`
	Disabled bool      `json:"disabled"`
	Last     time.Time `json:"last"`
	Created  time.Time `json:"created" storm:"created"`
	//Protocol *Protocol `json:"protocol"`
}

//LinkEvent 链接历史
type LinkEvent struct {
	Id      int       `json:"id" storm:"id,increment"`
	LinkId  int       `json:"link_id" storm:"index"`
	Event   string    `json:"event"`
	Created time.Time `json:"created" storm:"created"`
}
