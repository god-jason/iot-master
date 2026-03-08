package iot_master

import (
	"embed"

	"github.com/god-jason/iot-master/iot"
	"github.com/god-jason/iot-master/pkg/menu"
	"github.com/god-jason/iot-master/pkg/page"
	"github.com/god-jason/iot-master/pkg/store"
	"github.com/god-jason/iot-master/pkg/web"
)

//go:embed menu.json
var menuJson []byte

//go:embed pages
var pages embed.FS

//go:embed protocols
var protocols embed.FS

//go:embed dist/browser
var www embed.FS

func init() {

	page.PagesFS = store.PrefixFS(&pages, "pages")

	err := menu.Content(menuJson)
	if err != nil {
		panic(err)
	}

	//加载协议
	iot.Protocols = store.PrefixFS(&protocols, "protocols")

	//前端页面
	web.StaticFS(www, "/", "dist/browser", "index.html")
}
