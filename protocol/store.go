package protocol

import (
	"embed"
	"github.com/busy-cloud/boat/store"
)

var protocolsStore store.Store

func Dir(dir string, base string) {
	protocolsStore.Dir(dir, base)
}

func Zip(zip string, base string) {
	protocolsStore.Zip(zip, base)
}

func EmbedFS(fs embed.FS, base string) {
	protocolsStore.EmbedFS(fs, base)
}
