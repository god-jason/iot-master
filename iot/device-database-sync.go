package iot

import (
	"time"

	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/boat/table"
)

func databaseSync(id string, databases map[string]map[string]Sync) (has bool, err error) {
	hasSync := false

	//同步连接
	links, ok := databases["link"]
	if ok {
		has, err = databaseSyncLinks(id, links)
		if err != nil {
			return
		}
		if has {
			hasSync = true
		}
	}

	//同步设备
	devices, ok := databases["device"]
	if ok {
		has, err = databaseSyncDevices(id, devices)
		if err != nil {
			return
		}
		if has {
			hasSync = true
		}
	}

	//同步物模型

	return hasSync, nil
}

func databaseSyncLinks(id string, links map[string]Sync) (has bool, err error) {
	tab, err := table.Get("link")
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

	//不能删除默认连接
	if len(rows) == 0 {
		return false, nil
	}

	//数量不匹配，全部更新
	if len(rows) != len(links) {
		mqtt.Publish("device/"+id+"/database/link/clear", nil)
		mqtt.Publish("device/"+id+"/database/link/insertArray", rows)
		return true, nil
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
			return true, nil
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
		mqtt.Publish("device/"+id+"/database/link/clear", nil)
		mqtt.Publish("device/"+id+"/database/link/insertArray", rows)
	}

	return has, nil
}

func databaseSyncDevices(id string, devices map[string]Sync) (has bool, err error) {
	tab, err := table.Get("device")
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
	if len(rows) != len(devices) {
		mqtt.Publish("device/"+id+"/database/device/clear", nil)
		mqtt.Publish("device/"+id+"/database/device/insertArray", rows)
		return true, nil
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
			return true, nil
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
		mqtt.Publish("device/"+id+"/database/device/clear", nil)
		mqtt.Publish("device/"+id+"/database/device/insertArray", rows)
	}

	return
}

func databaseSyncModels(id string, models map[string]Sync) (has bool, err error) {
	tab, err := table.Get("device")
	if err != nil {
		return false, err
	}

	//先查出所有需要的产品ID
	rows, err := tab.Find(&table.ParamSearch{
		Skip:   0,
		Limit:  999,
		Filter: map[string]any{"gateway_id": id},
		Fields: []string{"id", "gateway_id", "product_id"},
	})
	var pids map[string]any
	for _, row := range rows {
		pids[row["product_id"].(string)] = nil
	}
	if len(rows) == 0 {
		return false, nil
	}

	tab, err = table.Get("product_model")
	if err != nil {
		log.Error(err)
		return
	}

	//查出所有依赖的物模型
	rows = nil
	for pid, _ := range pids {
		row, err := tab.Get(pid, nil)
		if err != nil {
			return false, err
		}
		rows = append(rows, row)
	}

	//不能删除默认模型
	if len(rows) == 0 {
		return false, nil
	}

	//数量不匹配，全部更新
	if len(rows) != len(models) {
		mqtt.Publish("device/"+id+"/database/model/clear", nil)
		mqtt.Publish("device/"+id+"/database/model/insertArray", rows)
		return true, nil
	}

	//逐一比较日期
	for i, m := range models {
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
		mqtt.Publish("device/"+id+"/database/model/clear", nil)
		mqtt.Publish("device/"+id+"/database/model/insertArray", rows)
	}

	return has, nil
}
