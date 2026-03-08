package web

import (
	"mime"

	"github.com/busy-cloud/boat/log"
)

func init() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Error(err)
	}
}
