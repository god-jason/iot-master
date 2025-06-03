package iot_master

import (
	"embed"
	"encoding/json"
	_ "github.com/busy-cloud/boat-ui"
	"github.com/busy-cloud/boat/apps"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/store"
	_ "github.com/god-jason/iot-master/internal"
	"github.com/god-jason/iot-master/protocol"
)

//go:embed assets
var assets embed.FS

//go:embed pages
var pages embed.FS

//go:embed manifest.json
var manifest []byte

//go:embed protocols
var protocols embed.FS

func init() {
	//注册协议
	protocol.EmbedFS(&protocols, "protocols")

	//注册为内部插件
	var a apps.App
	err := json.Unmarshal(manifest, &a)
	if err != nil {
		log.Fatal(err)
	}
	apps.Register(&a)

	//注册资源
	a.AssetsFS = store.PrefixFS(&assets, "assets")
	a.PagesFS = store.PrefixFS(&pages, "pages")
}
