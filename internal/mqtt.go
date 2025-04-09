package internal

import (
	"encoding/json"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"strings"
)

func Startup() error {
	mqtt.Subscribe("device/+/+/property", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		id := ss[2]
		var values map[string]any
		err := json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}

		d := devices.Load(id)
		if d == nil {
			d = &Device{}
			has, err := db.Engine().ID(id).Get(&d.Device)
			if err != nil {
				log.Error(err)
				return
			}
			if !has {
				log.Error("device not exist")
				return
			}
			err = d.Open()
			if err != nil {
				log.Error(err)
			}

			devices.Store(id, d)
		}

		d.PutValues(values)
	})

	return nil
}
