package iot_master

import (
	"embed"
	_ "github.com/busy-cloud/boat"
	_ "github.com/busy-cloud/boat-ui"
	"github.com/busy-cloud/boat/menu"
	"github.com/busy-cloud/boat/page"
	_ "github.com/god-jason/iot-master/internal"
	"github.com/god-jason/iot-master/protocol"
)

//go:embed pages
var pages embed.FS

//go:embed menus
var menus embed.FS

//go:embed protocols
var protocols embed.FS

func init() {
	//注册页面
	page.EmbedFS(pages, "pages")

	//注册菜单
	menu.EmbedFS(menus, "menus")

	//注册协议
	protocol.EmbedFS(protocols, "protocols")
}
