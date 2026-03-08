package web

import (
	"mime"

	"github.com/god-jason/iot-master/pkg/log"
)

func init() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Error(err)
	}
}
