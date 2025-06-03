package protocol

import (
	"embed"
	"github.com/busy-cloud/boat/store"
)

var protocolsStore store.Store

func Dir(dir string) {
	protocolsStore.AddDir(dir)
}

func Zip(zip string) {
	protocolsStore.AddZip(zip)
}

func EmbedFS(fs *embed.FS, base string) {
	protocolsStore.Add(store.PrefixFS(fs, base))
}
