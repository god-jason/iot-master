package menu

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/god-jason/iot-master/pkg/api"
)

type Item struct {
	Name       string   `json:"name"`
	Title      string   `json:"title,omitempty"`
	Icon       string   `json:"icon,omitempty"`
	Url        string   `json:"url,omitempty"`
	External   bool     `json:"external,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
}

type Menu struct {
	Name       string   `json:"name"`
	Title      string   `json:"title,omitempty"`
	NzIcon     string   `json:"nz_icon,omitempty"` //ant.design图标库
	Items      []*Item  `json:"items,omitempty"`
	Index      int      `json:"index,omitempty"`
	Privileges []string `json:"privileges,omitempty"`
	Admin      bool     `json:"admin,omitempty"` //管理员
	//Domain     []string `json:"domain"` //域 admin project 或 dealer等
}

var menus []Menu
var filename string = "menu.json" //默认menu.json

func init() {
	api.RegisterUnAuthorized("GET", "menu", func(ctx *gin.Context) {
		if menus != nil {
			ctx.JSON(200, menus)
			return
		}
		ctx.File(filename)
	})
}

func Content(buf []byte) error {
	return json.Unmarshal(buf, &menus)
}

func File(fn string) {
	filename = fn
}
