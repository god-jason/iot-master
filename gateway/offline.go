package gateway

import (
	"fmt"
	"github.com/god-jason/iot-master/lib"
	"github.com/god-jason/iot-master/log"
	"github.com/god-jason/iot-master/mqtt"
	"time"
)

var offlineTimers lib.Map[time.Timer]

func Offline(pid, id string) {
	timer := offlineTimers.Load(id)
	if timer != nil {
		return
	}

	//延迟1分钟报警
	timer = time.AfterFunc(time.Minute, func() {
		topic := fmt.Sprintf("offline/%s/%s", pid, id)
		err := mqtt.Publish(topic, nil)
		if err != nil {
			log.Error(err)
		}
		offlineTimers.Delete(id)
	})
	offlineTimers.Store(id, timer)
}

func Online(pid, id string) {
	timer := offlineTimers.Load(id)
	if timer != nil {
		//如果有延迟报警，删除之
		timer.Stop()
		offlineTimers.Delete(id)
		return
	}

	topic := fmt.Sprintf("online/%s/%s", pid, id)
	err := mqtt.Publish(topic, nil)
	if err != nil {
		log.Error(err)
	}
}
