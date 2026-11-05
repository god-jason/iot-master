package memory_test

import (
	"testing"
	"time"

	"github.com/god-jason/iot-master/pkg/tsdb"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

func TestMemoryStoreWriteQueryLast(t *testing.T) {
	s, err := tsdb.New("memory", nil)
	if err != nil {
		t.Fatalf("New(memory): %v", err)
	}

	// Write 2 个设备各 2 条
	now := time.Now().Truncate(time.Second)
	err = s.Write(nil, tsdbapi.Point{
		Measurement: "device_data",
		Tags:        map[string]string{"id": "dev1"},
		Fields:      map[string]any{"temperature": 25.5, "humidity": 60.0},
		Time:        now,
	})
	if err != nil {
		t.Fatalf("Write dev1 @now: %v", err)
	}
	err = s.Write(nil, tsdbapi.Point{
		Measurement: "device_data",
		Tags:        map[string]string{"id": "dev1"},
		Fields:      map[string]any{"temperature": 26.5, "humidity": 61.0},
		Time:        now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("Write dev1 @now+1m: %v", err)
	}
	err = s.Write(nil, tsdbapi.Point{
		Measurement: "device_data",
		Tags:        map[string]string{"id": "dev2"},
		Fields:      map[string]any{"temperature": 30.0},
		Time:        now,
	})
	if err != nil {
		t.Fatalf("Write dev2: %v", err)
	}

	// Query dev1（窗口 1s 不聚合，原始数据）
	series, err := s.Query(nil, tsdbapi.QuerySpec{
		Measurement: "device_data",
		Tags:        map[string]string{"id": "dev1"},
		Start:       now.Add(-time.Second),
		End:         now.Add(2 * time.Minute),
		Fields:      []string{"temperature"},
		Window:      "1s",
		Method:      "mean",
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(series["temperature"].Values) != 2 {
		t.Errorf("expected 2 temperature values, got %d: %+v", len(series["temperature"].Values), series["temperature"].Values)
	}

	// Last dev1
	lasts, err := s.Last(nil, "device_data", map[string]string{"id": "dev1"})
	if err != nil {
		t.Fatalf("Last: %v", err)
	}
	if v, ok := lasts["temperature"]; !ok {
		t.Errorf("expected temperature last, got %+v", lasts)
	} else if v.Value != 26.5 {
		t.Errorf("expected last temperature 26.5, got %v", v.Value)
	}

	// dev1 不应该查到 dev2
	series2, _ := s.Query(nil, tsdbapi.QuerySpec{
		Measurement: "device_data",
		Tags:        map[string]string{"id": "dev2"},
		Start:       now,
		End:         now.Add(time.Minute),
	})
	if _, ok := series2["temperature"]; !ok {
		t.Errorf("dev2 temperature should exist: %+v", series2)
	}

	// Close
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
