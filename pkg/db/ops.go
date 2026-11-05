package db

import (
	"time"

	"github.com/bytedance/sonic"
	"xorm.io/builder"
)

// Event 设备事件记录表
type Event struct {
	ID        string    `xorm:"id pk varchar(24)" json:"id"`
	PK        string    `xorm:"pk index" json:"pk"`
	DN        string    `xorm:"dn" json:"dn"`
	Type      string    `xorm:"type" json:"type"`
	Code      string    `xorm:"code" json:"code"`
	Message   string    `xorm:"message" json:"message"`
	Data      string    `xorm:"data" json:"data"`
	Timestamp int64     `xorm:"timestamp" json:"timestamp"`
	Created   time.Time `xorm:"created" json:"created,omitempty"`
}

// Log 设备日志记录表
type Log struct {
	ID        string    `xorm:"id pk varchar(24)" json:"id"`
	PK        string    `xorm:"pk index" json:"pk"`
	DN        string    `xorm:"dn" json:"dn"`
	Level     string    `xorm:"level" json:"level"`
	Message   string    `xorm:"message" json:"message"`
	Data      string    `xorm:"data" json:"data"`
	Timestamp int64     `xorm:"timestamp" json:"timestamp"`
	Created   time.Time `xorm:"created" json:"created,omitempty"`
}

// Config 设备运行时配置表
type Config struct {
	ID        string    `xorm:"id pk varchar(24)" json:"id"`
	PK        string    `xorm:"pk index" json:"pk"`
	DN        string    `xorm:"dn" json:"dn"`
	Key       string    `xorm:"key" json:"key"`
	Value     string    `xorm:"value" json:"value"`
	UpdatedAt int64     `xorm:"updated_at" json:"updated_at"`
	Created   time.Time `xorm:"created" json:"created,omitempty"`
}

func marshalData(data interface{}) string {
	if data == nil {
		return ""
	}
	if s, ok := data.(string); ok {
		return s
	}
	b, err := sonic.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}

// SaveEvent 保存设备事件
func SaveEvent(pk, dn, eventType, code, message string, data interface{}, timestamp int64) error {
	doc := map[string]interface{}{
		"id":        NewID(),
		"pk":        pk,
		"dn":        dn,
		"type":      eventType,
		"code":      code,
		"message":   message,
		"data":      marshalData(data),
		"timestamp": timestamp,
	}
	_, err := InsertOne("events", doc)
	return err
}

// SaveLog 保存设备日志
func SaveLog(pk, dn, level, message string, data interface{}, timestamp int64) error {
	doc := map[string]interface{}{
		"id":        NewID(),
		"pk":        pk,
		"dn":        dn,
		"level":     level,
		"message":   message,
		"data":      marshalData(data),
		"timestamp": timestamp,
	}
	_, err := InsertOne("device_logs", doc)
	return err
}

// SaveConfig 保存/更新设备配置（upsert）
func SaveConfig(pk, dn, key string, value interface{}, updatedAt int64) error {
	cond := builder.Eq{"pk": pk, "dn": dn, "key": key}
	create := map[string]interface{}{
		"id":         NewID(),
		"pk":         pk,
		"dn":         dn,
		"key":        key,
		"value":      marshalData(value),
		"updated_at": updatedAt,
	}
	update := map[string]interface{}{
		"value":      marshalData(value),
		"updated_at": updatedAt,
	}
	_, err := Upsert("device_configs", cond, create, update)
	return err
}

// GetConfig 获取单条设备配置
func GetConfig(pk, dn, key string) (*Config, error) {
	var c Config
	cond := builder.Eq{"pk": pk, "dn": dn, "key": key}
	found, err := FindOne("device_configs", cond, &c)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &c, nil
}

// GetDeviceConfigs 获取设备全部配置
func GetDeviceConfigs(pk, dn string) ([]Config, error) {
	var configs []Config
	cond := builder.Eq{"pk": pk, "dn": dn}
	err := Find("device_configs", cond, &configs, 0, 0)
	return configs, err
}

// GetDeviceEvents 按时间范围查询设备事件
func GetDeviceEvents(pk, dn string, startTime, endTime int64) ([]Event, error) {
	cond := builder.And(
		builder.Eq{"pk": pk, "dn": dn},
		builder.Gte{"timestamp": startTime},
		builder.Lte{"timestamp": endTime},
	)
	var events []Event
	err := Find("events", cond, &events, 0, 0)
	return events, err
}

// GetDeviceLogs 按时间范围查询设备日志
func GetDeviceLogs(pk, dn string, startTime, endTime int64) ([]Log, error) {
	cond := builder.And(
		builder.Eq{"pk": pk, "dn": dn},
		builder.Gte{"timestamp": startTime},
		builder.Lte{"timestamp": endTime},
	)
	var logs []Log
	err := Find("device_logs", cond, &logs, 0, 0)
	return logs, err
}
