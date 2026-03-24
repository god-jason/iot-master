package iot

import (
	"fmt"
	"time"

	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/mqtt"
	"github.com/god-jason/iot-master/pkg/table"
	"xorm.io/xorm/schemas"
)

func deviceSettingSync(id string, sts map[string]int) (has bool, err error) {
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
			//读取配置，然后保存
			topic := fmt.Sprintf("device/%s/setting/%s/read", id, setting.Name)
			mqtt.Publish(topic, nil)
		}
	}

	return
}

func modelSync(id string, sts map[string]int) (has bool, err error) {
	for pid, ver := range sts {
		var setting ProductSetting
		has, err := db.Engine().ID(schemas.PK{pid, "model"}).Get(&setting)
		if err == nil && has {
			if ver < setting.Version {
				topic := fmt.Sprintf("device/%s/database/model/insert", id)
				mqtt.Publish(topic, &setting)
				has = true
			}
		}
	}
	return
}

func databaseSync(id string, name string, records map[string]Sync) (has bool, err error) {
	tab, err := table.Get("name")
	if err != nil {
		return false, err
	}

	//查找子设备
	rows, err := tab.Find(&table.ParamSearch{
		Skip:   0,
		Limit:  999,
		Filter: map[string]any{"gateway_id": id},
	})
	if err != nil {
		return false, err
	}

	//数量不匹配，全部更新
	if len(rows) != len(records) {
		mqtt.Publish("device/"+id+"/database/"+name+"/clear", nil)
		mqtt.Publish("device/"+id+"/database/"+name+"/insertArray", rows)
		return true, nil
	}

	//ID不匹配，全部更新
	for i, _ := range records {
		found := false
		for _, r := range rows {
			if r["id"] == i {
				found = true
				break
			}
		}
		if !found {
			mqtt.Publish("device/"+id+"/database/"+name+"/clear", nil)
			mqtt.Publish("device/"+id+"/database/"+name+"/insertArray", rows)
			return true, nil
		}
	}

	//麻烦的逐一同步
	for i, m := range records {
		//找到对应的row
		var row map[string]any
		for _, r := range rows {
			if r["id"] == i {
				row = r
				break
			}
		}

		//检查更新时间
		if u, ok := row["updated"]; ok {
			if t, ok := u.(time.Time); ok {
				tt, err := time.Parse(time.DateTime, m.Updated)
				if err != nil || tt.Before(t) {
					has = true
					break
				}
			}
		}

		//检查创建时间
		if u, ok := row["created"]; ok {
			if t, ok := u.(time.Time); ok {
				tt, err := time.Parse(time.DateTime, m.Created)
				if err != nil || tt.Before(t) {
					has = true
					break
				}
			}
		}
	}

	if has {
		mqtt.Publish("device/"+id+"/database/"+name+"/clear", nil)
		mqtt.Publish("device/"+id+"/database/"+name+"/insertArray", rows)
	}

	return
}
