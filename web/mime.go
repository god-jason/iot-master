package web

import (
	"github.com/zgwit/iot-master/log"
	"mime"
)

func init() {
	err := mime.AddExtensionType(".js", "application/javascript")
	if err != nil {
		log.Error(err)
	}
}
