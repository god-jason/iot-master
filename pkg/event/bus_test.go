package event

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestBusMatch(t *testing.T) {
	bus := NewEventBus()

	tests := []struct {
		filter string
		topic  string
		want   bool
	}{
		{"a/b/c", "a/b/c", true},
		{"a/b/c", "a/b/d", false},
		{"a/b", "a/b/c", false},
		{"a/+/c", "a/b/c", true},
		{"a/+/c", "a/x/c", true},
		{"a/+/c", "a/b/d", false},
		{"a/#", "a/b/c", true},
		{"a/#", "a/x/y/z", true},
		{"#", "any/thing/goes", true},
		{"device/+/property", "device/dev01/property", true},
		{"device/+/property", "device/dev01/event", false},
		{"device/+/property", "device/group/dev01/property", false},
		{"device/+/property/+", "device/dev01/property/temp", true},
		{"device/#", "device/dev01", true},
		{"device/#", "device", false},
	}

	for _, c := range tests {
		var hit int32
		unsub := bus.Subscribe(c.filter, func(topic string, payload any) {
			atomic.AddInt32(&hit, 1)
		})
		bus.Publish(c.topic, nil)
		time.Sleep(20 * time.Millisecond) // pool 异步投递
		unsub()

		got := atomic.LoadInt32(&hit) > 0
		if got != c.want {
			t.Errorf("filter=%q topic=%q → matched=%v want=%v", c.filter, c.topic, got, c.want)
		}
	}
}

func TestBusUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	var hit int32

	unsub := bus.Subscribe("device/+/property", func(topic string, payload any) {
		atomic.AddInt32(&hit, 1)
	})

	bus.Publish("device/dev01/property", nil)
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&hit) != 1 {
		t.Fatal("should hit once")
	}

	unsub()
	bus.Publish("device/dev01/property", nil)
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&hit) != 1 {
		t.Fatal("after unsubscribe should not hit")
	}
}

func TestBusMultipleHandlers(t *testing.T) {
	bus := NewEventBus()
	var total int32

	for i := 0; i < 5; i++ {
		bus.Subscribe("device/#", func(topic string, payload any) {
			atomic.AddInt32(&total, 1)
		})
	}
	bus.Subscribe("device/dev01/property", func(topic string, payload any) {
		atomic.AddInt32(&total, 1)
	})

	bus.Publish("device/dev01/property", nil)
	time.Sleep(30 * time.Millisecond)

	if got := atomic.LoadInt32(&total); got != 6 {
		t.Errorf("total=%d want 6 (5 # + 1 exact)", got)
	}
}

func TestBusAsync(t *testing.T) {
	bus := NewEventBus()
	var hit int32

	bus.Subscribe("a/#", func(topic string, payload any) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&hit, 1)
	})
	bus.Subscribe("a/#", func(topic string, payload any) {
		atomic.AddInt32(&hit, 1)
	})

	bus.Publish("a/b/c", nil)
	time.Sleep(40 * time.Millisecond)

	if got := atomic.LoadInt32(&hit); got != 2 {
		t.Errorf("async total=%d want 2", got)
	}
}

func TestBusTriePrune(t *testing.T) {
	bus := NewEventBus()
	h1 := func(topic string, payload any) {}
	h2 := func(topic string, payload any) {}

	u1 := bus.Subscribe("a/b/c", h1)
	bus.Subscribe("a/b/c", h2)
	if bus.Len() != 2 {
		t.Fatalf("len=%d want 2", bus.Len())
	}

	u1()
	if bus.Len() != 1 {
		t.Fatalf("len=%d want 1", bus.Len())
	}

	bus.Reset()
	if bus.Len() != 0 {
		t.Fatalf("len=%d want 0 after reset", bus.Len())
	}
}

func BenchmarkBusPublish(b *testing.B) {
	bus := NewEventBus()
	for i := 0; i < 1000; i++ {
		filter := "device/" + itoa(i%10) + "/+/property"
		bus.Subscribe(filter, func(topic string, payload any) {})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("device/1/dev01/property", nil)
	}
}

// itoa 内部辅助（供 bus_test 和 simple_test 共用）
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
