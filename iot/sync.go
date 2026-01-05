package iot

import (
	"time"

	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/boat/table"
)

func syncDevices(id string, devices map[string]Sync, models map[string]Sync) {
	tab, err := table.Get("device")
	if err != nil {
		log.Error(err)
		return
	}
	model, err := table.Get("product_model")
	if err != nil {
		log.Error(err)
		return
	}

	//查找子设备
	rows, err := tab.Find(&table.ParamSearch{
		Skip:   0,
		Limit:  999,
		Filter: map[string]any{"gateway_id": id},
	})
	if err != nil {
		log.Error(err)
		return
	}

	//数量不匹配，全部更新
	if len(rows) != len(devices) {
		mqtt.Publish("device/"+id+"/database/device/clear", nil)
		mqtt.Publish("device/"+id+"/database/device/insertArray", rows)
		return
	}

	//ID不匹配，全部更新
	for i, _ := range devices {
		found := false
		for _, r := range rows {
			if r["id"] == i {
				found = true
				break
			}
		}
		if !found {
			mqtt.Publish("device/"+id+"/database/device/clear", nil)
			mqtt.Publish("device/"+id+"/database/device/insertArray", rows)
			break
		}
	}

	//麻烦的逐一同步
	for i, m := range devices {
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
					mqtt.Publish("device/"+id+"/database/device/insert", row)
					continue
				}
			}
		}

		//检查创建时间
		if u, ok := row["created"]; ok {
			if t, ok := u.(time.Time); ok {
				tt, err := time.Parse(time.DateTime, m.Created)
				if err != nil || tt.Before(t) {
					mqtt.Publish("device/"+id+"/database/device/insert", row)
					continue
				}
			}
		}
	}

	//同步设备所需的物模型
	if len(models) == 0 {
		for _, row := range rows {
			pid := row["product_id"].(string)
			row, err := model.Get(pid, nil)
			if err != nil {
				log.Error(err)
				continue
			}
			mqtt.Publish("device/"+id+"/database/model/insert", row)
		}
	} else {
		for _, row := range rows {
			pid := row["product_id"].(string)
			if _, ok := models[pid]; !ok {
				row, err := model.Get(pid, nil)
				if err != nil {
					log.Error(err)
					continue
				}
				mqtt.Publish("device/"+id+"/database/model/insert", row)
			}
		}
	}
}

func syncLinks(id string, links map[string]Sync) {
	tab, err := table.Get("link")
	if err != nil {
		log.Error(err)
		return
	}
	//查找子设备
	rows, err := tab.Find(&table.ParamSearch{
		Skip:   0,
		Limit:  999,
		Filter: map[string]any{"gateway_id": id},
	})
	if err != nil {
		log.Error(err)
		return
	}

	//不能删除默认连接
	if len(rows) == 0 {
		return
	}

	//数量不匹配，全部更新
	if len(rows) != len(links) {
		mqtt.Publish("device/"+id+"/database/link/clear", nil)
		mqtt.Publish("device/"+id+"/database/link/insertArray", rows)
		return
	}

	//ID不匹配，全部更新
	for i, _ := range links {
		found := false
		for _, r := range rows {
			if r["id"] == i {
				found = true
				break
			}
		}
		if !found {
			mqtt.Publish("device/"+id+"/database/link/clear", nil)
			mqtt.Publish("device/"+id+"/database/link/insertArray", rows)
			break
		}
	}

	//麻烦的逐一同步
	for i, m := range links {
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
					mqtt.Publish("device/"+id+"/database/link/insert", row)
					continue
				}
			}
		}

		//检查创建时间
		if u, ok := row["created"]; ok {
			if t, ok := u.(time.Time); ok {
				tt, err := time.Parse(time.DateTime, m.Created)
				if err != nil || tt.Before(t) {
					mqtt.Publish("device/"+id+"/database/link/insert", row)
					continue
				}
			}
		}
	}
}

func syncModels(id string, models map[string]Sync) {
	tab, err := table.Get("product_model")
	if err != nil {
		log.Error(err)
		return
	}

	for i, m := range models {
		row, err := tab.Get(i, nil)
		if err != nil {
			log.Error(err)
			continue
		}

		//检查更新时间
		if u, ok := row["updated"]; ok {
			if t, ok := u.(time.Time); ok {
				tt, err := time.Parse(time.DateTime, m.Updated)
				if err != nil || tt.Before(t) {
					mqtt.Publish("device/"+id+"/database/model/insert", row)
					continue
				}
			}
		}

		//检查创建时间
		if u, ok := row["created"]; ok {
			if t, ok := u.(time.Time); ok {
				tt, err := time.Parse(time.DateTime, m.Created)
				if err != nil || tt.Before(t) {
					mqtt.Publish("device/"+id+"/database/model/insert", row)
					continue
				}
			}
		}
	}
}
