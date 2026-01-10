package iot

import (
	"fmt"

	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/mqtt"
)

func settingSync(id string, sts map[string]int) (has bool, err error) {
	//查出所有配置文件
	var settings []DeviceSetting
	err = db.Engine().Where("id=?", id).Find(&settings)
	if err != nil {
		return false, err
	}

	//同步配置文件
	for _, setting := range settings {
		if ver, has := sts[setting.Name]; !has || ver < setting.Version {
			topic := fmt.Sprintf("device/%s/setting", id)
			mqtt.Publish(topic, &setting)
			has = true
		} else if ver > setting.Version {
			//读取配置
			topic := fmt.Sprintf("device/%s/setting/%s/read", id, setting.Name)
			mqtt.Publish(topic, nil)
		}
	}

	return
}
