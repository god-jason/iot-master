package history

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/god-jason/iot-master/pkg/log"
	"github.com/god-jason/iot-master/pkg/mqtt"
)

func subscribe() {
	mqtt.Subscribe("history/+/+/values", func(topic string, payload []byte) {
		var values map[string]interface{}
		err := json.Unmarshal(payload, &values)
		if err != nil {
			log.Error(err)
			return
		}

		ss := strings.Split(topic, "/")
		pid := ss[1]
		id := ss[2]

		err = Write(pid, id, time.Now().UnixMilli(), values)
		if err != nil {
			log.Error(err)
		}
	})
}
