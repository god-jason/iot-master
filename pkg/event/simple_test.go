package event

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSimpleBasic(t *testing.T) {
	bus := NewSimple()
	var hit int32

	bus.Subscribe("device/dev01/property", func(topic string, payload any) {
		atomic.AddInt32(&hit, 1)
	})

	bus.Publish("device/dev01/property", nil)
	bus.Publish("device/dev01/event", nil) // 不同 topic，不应该命中
	time.Sleep(20 * time.Millisecond)

	if got := atomic.LoadInt32(&hit); got != 1 {
		t.Errorf("hit=%d want 1", got)
	}
}

func TestSimpleUnsubscribe(t *testing.T) {
	bus := NewSimple()
	var hit int32

	unsub := bus.Subscribe("x/y", func(topic string, payload any) {
		atomic.AddInt32(&hit, 1)
	})

	bus.Publish("x/y", nil)
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&hit) != 1 {
		t.Fatal("should hit once")
	}

	unsub()
	bus.Publish("x/y", nil)
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&hit) != 1 {
		t.Fatal("after unsubscribe should not hit")
	}
}

func TestSimpleMultipleOnSameTopic(t *testing.T) {
	bus := NewSimple()
	var total int32

	for i := 0; i < 3; i++ {
		bus.Subscribe("a/b", func(topic string, payload any) {
			atomic.AddInt32(&total, 1)
		})
	}

	bus.Publish("a/b", nil)
	time.Sleep(20 * time.Millisecond)
	if got := atomic.LoadInt32(&total); got != 3 {
		t.Errorf("total=%d want 3", got)
	}
}

func TestSimpleAsync(t *testing.T) {
	bus := NewSimple()
	var hit int32

	bus.Subscribe("a/b", func(topic string, payload any) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&hit, 1)
	})

	bus.Publish("a/b", nil)
	time.Sleep(30 * time.Millisecond)

	if got := atomic.LoadInt32(&hit); got != 1 {
		t.Errorf("async hit=%d want 1", got)
	}
}

func TestSimpleReset(t *testing.T) {
	bus := NewSimple()
	bus.Subscribe("a/b", func(topic string, payload any) {})
	bus.Reset()

	var hit int32
	bus.Publish("a/b", nil)
	time.Sleep(20 * time.Millisecond)
	if atomic.LoadInt32(&hit) != 0 {
		t.Fatal("after reset should not hit")
	}
}

func BenchmarkSimplePublish(b *testing.B) {
	bus := NewSimple()
	for i := 0; i < 1000; i++ {
		topic := "device/" + itoa(i) + "/property"
		bus.Subscribe(topic, func(topic string, payload any) {})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish("device/500/property", nil)
	}
}
