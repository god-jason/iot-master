package web

import (
	"github.com/god-jason/iot-master/log"
	"mime"
)

func init() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Error(err)
	}
}
