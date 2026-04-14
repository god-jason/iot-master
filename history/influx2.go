package history

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/god-jason/iot-master/pkg/config"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/spf13/cast"
)

var client influxdb2.Client
var writer api.WriteAPIBlocking
var reader api.QueryAPI

type Point struct {
	Value any   `json:"value"`
	Time  int64 `json:"time"`
}

func Startup() error {
	if !config.GetBool(MODULE, "enable") {
		return nil
	}

	client = influxdb2.NewClient(config.GetString(MODULE, "url"), config.GetString(MODULE, "token"))
	writer = client.WriteAPIBlocking(config.GetString(MODULE, "org"), config.GetString(MODULE, "bucket"))
	reader = client.QueryAPI(config.GetString(MODULE, "org"))

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
	if writer == nil {
		return nil
	}

	vs := make(map[string]any)
	for k, v := range values {
		k = strings.TrimSpace(k)

		//过滤无效字段名
		if k == "" {
			continue
		}
		//处理数据类型
		val, err := cast.ToFloat64E(v)
		if err == nil {
			vs[k] = val
		}
	}
	if len(vs) == 0 {
		return nil
	}

	return writer.WritePoint(context.Background(), write.NewPoint(table, map[string]string{"id": id}, vs, time.UnixMilli(timestamp)))
}

func Query(table, id, name, start, end, window, method string) ([]*Point, error) {
	if reader == nil {
		return nil, errors.New("influxdb未启用")
	}

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
