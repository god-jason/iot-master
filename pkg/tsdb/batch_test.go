package tsdb

import (
	"testing"
	"time"

	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

func TestWriteUsesBatchAsync(t *testing.T) {
	store, err := New("memory", nil)
	if err != nil {
		t.Fatalf("New(memory): %v", err)
	}
	DefaultStore = store
	defer func() { DefaultStore = nil }()

	now := time.Now()
	for i := 0; i < 3; i++ {
		Write(tsdbapi.Point{
			Measurement: "dev_data",
			Tags:        map[string]string{"id": "dev1"},
			Fields:      map[string]any{"temperature": float64(i)},
			Time:        now.Add(time.Duration(i) * time.Second),
		})
	}

	// Write 是异步的，最多等 2s 观察落库
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		lasts, err := store.Last(nil, "dev_data", map[string]string{"id": "dev1"})
		if err != nil {
			t.Fatalf("Last: %v", err)
		}
		if v, ok := lasts["temperature"]; ok {
			if v.Value != float64(2) {
				t.Errorf("expected last temperature 2, got %v", v.Value)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("tsdb 批量写入 2s 内未落库")
}
