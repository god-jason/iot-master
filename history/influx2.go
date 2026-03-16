package history

import (
	"context"
	"time"

	"github.com/god-jason/iot-master/pkg/config"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

var client influxdb2.Client
var writer api.WriteAPI
var reader api.QueryAPI

type Point struct {
	Value any   `json:"value"`
	Time  int64 `json:"time"`
}

func Startup() error {
	client = influxdb2.NewClient(config.GetString(MODULE, "url"), config.GetString(MODULE, "token"))
	writer = client.WriteAPI(config.GetString(MODULE, "org"), config.GetString(MODULE, "bucket"))
	reader = client.QueryAPI(config.GetString(MODULE, "org"))

	//订阅消息
	subscribe()

	return nil
}

func Shutdown() error {
	client.Close()
	return nil
}

func Client() influxdb2.Client {
	return client
}

func Write(table, id string, timestamp int64, values map[string]any) error {
	writer.WritePoint(write.NewPoint(table, map[string]string{"id": id}, values, time.UnixMilli(timestamp)))
	return nil
}

func Query(table, id, name, start, end, window, method string) ([]*Point, error) {
	bucket := config.GetString(MODULE, "bucket")

	flux := "from(bucket: \"" + bucket + "\")\n"
	flux += "|> range(start: " + start + ", stop: " + end + ")\n"
	flux += "|> filter(fn: (r) => r[\"_measurement\"] == \"" + table + "\")\n"
	flux += "|> filter(fn: (r) => r[\"id\"] == \"" + id + "\")\n"
	flux += "|> filter(fn: (r) => r[\"_field\"] == \"" + name + "\")"
	flux += "|> aggregateWindow(every: " + window + ", fn: " + method + ", createEmpty: false)\n"
	flux += "|> yield(name: \"" + method + "\")"

	result, err := reader.Query(context.Background(), flux)
	if err != nil {
		return nil, err
	}

	var records []*Point
	for result.Next() {
		//result.TableChanged() 查询多个数值的情况？？？
		records = append(records, &Point{
			Value: result.Record().Value(),
			Time:  result.Record().Time().UnixMilli(),
		})
	}
	return records, result.Err()
}
