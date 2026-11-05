package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExpiredCache_Basic(t *testing.T) {
	c := NewExpiredCache[string, int](func(key string) (int, error) {
		return len(key), nil
	}, time.Second)

	v, err := c.Get("abc")
	if err != nil || v != 3 {
		t.Fatalf("miss: got %d %v, want 3 <nil>", v, err)
	}

	v, err = c.Get("abc")
	if err != nil || v != 3 {
		t.Fatalf("hit: got %d %v", v, err)
	}
}

func TestExpiredCache_TTLExpiry(t *testing.T) {
	c := NewExpiredCache[string, int](func(key string) (int, error) {
		return 1, nil
	}, 50*time.Millisecond)

	_, _ = c.Get("k")
	if !c.Contains("k") {
		t.Fatal("should contain before expiry")
	}

	time.Sleep(80 * time.Millisecond)

	// Contains 过期后应返回 false
	if c.Contains("k") {
		t.Fatal("should NOT contain after expiry")
	}

	// Get 过期后走 loader（但我们没测 loader 次数，简单验证可用）
	// 直接用 Refresh 规避不确定时序
	c.Refresh("k")
}

func TestExpiredCache_PurgeExpired(t *testing.T) {
	c := NewExpiredCache[int, string](func(key int) (string, error) {
		return "", nil
	}, 30*time.Millisecond)

	c.Set(1, "a")
	c.Set(2, "b")

	time.Sleep(50 * time.Millisecond)

	if c.Len() != 2 {
		t.Fatalf("before Purge: Len=%d, want 2", c.Len())
	}
	c.PurgeExpired()
	if c.Len() != 0 {
		t.Fatalf("after Purge: Len=%d, want 0", c.Len())
	}
}

func TestExpiredCache_SetWithTTL(t *testing.T) {
	c := NewExpiredCache[string, int](nil, time.Hour) // 长 TTL 默认

	// 自定义短 TTL
	c.SetWithTTL("k", 99, 30*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	if c.Contains("k") {
		t.Fatal("SetWithTTL should expire independently from default TTL")
	}
}

func TestExpiredCache_SetTTL_NoRetroactive(t *testing.T) {
	// SetTTL 只影响新写入的条目
	c := NewExpiredCache[string, int](nil, 500*time.Millisecond)
	c.Set("old", 1)
	c.SetTTL(20 * time.Millisecond)
	c.Set("new", 2)

	time.Sleep(30 * time.Millisecond)

	if c.Contains("old") == false {
		t.Fatal("old entry should NOT be affected by SetTTL")
	}
	if c.Contains("new") {
		t.Fatal("new entry should expire under new TTL")
	}
}

func TestExpiredCache_ErrorCaching(t *testing.T) {
	fail := errors.New("boom")
	callCount := 0
	c := NewExpiredCache[string, int](func(key string) (int, error) {
		callCount++
		return 0, fail
	}, 50*time.Millisecond)

	_, _ = c.Get("k") // loader 失败，按 TTL 缓存
	if callCount != 1 {
		t.Fatalf("first call: loader=%d", callCount)
	}

	// TTL 内不重试
	_, _ = c.Get("k")
	if callCount != 1 {
		t.Fatalf("TTL内不应重试: loader=%d", callCount)
	}

	// 等过期后应自动重试
	time.Sleep(80 * time.Millisecond)
	_, _ = c.Get("k")
	if callCount != 2 {
		t.Fatalf("TTL过后应重试: loader=%d, want 2", callCount)
	}
}

func TestExpiredCache_SetErr_Expiry(t *testing.T) {
	myErr := errors.New("manual")
	c := NewExpiredCache[string, int](nil, 30*time.Millisecond)
	c.SetErr("k", myErr)

	if !c.Contains("k") {
		t.Fatal("should contain error-marked entry")
	}
	_, err := c.Get("k")
	if !errors.Is(err, myErr) {
		t.Fatalf("want myErr, got %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if c.Contains("k") {
		t.Fatal("error entry should expire")
	}
}

func TestExpiredCache_KeysFiltersExpired(t *testing.T) {
	c := NewExpiredCache[int, int](nil, 30*time.Millisecond)
	c.Set(1, 1)
	c.Set(2, 2)

	time.Sleep(50 * time.Millisecond)
	c.Set(3, 3)

	keys := c.Keys()
	if len(keys) != 1 || keys[0] != 3 {
		t.Fatalf("Keys() should return only unexpired, got %v", keys)
	}
}

func TestExpiredCache_LoaderNotSet(t *testing.T) {
	c := NewExpiredCache[string, int](nil, time.Second)
	_, err := c.Get("x")
	if !errors.Is(err, ErrLoaderNotSet) {
		t.Fatalf("want ErrLoaderNotSet, got %v", err)
	}
}

func TestExpiredCache_Singleflight(t *testing.T) {
	var calls int32
	c := NewExpiredCache[int, int](func(key int) (int, error) {
		atomic.AddInt32(&calls, 1)
		for i := 0; i < 1_000_000; i++ {
			_ = i
		}
		return key * 2, nil
	}, time.Second)

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			v, err := c.Get(7)
			if err != nil || v != 14 {
				t.Errorf("got %d %v", v, err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("singleflight broken: loader=%d, want 1", got)
	}
}

func TestExpiredCache_SingleflightExpiredKey(t *testing.T) {
	// 过期后的 key 被并发 Get，应重新 singleflight
	var calls int32
	c := NewExpiredCache[string, int](func(key string) (int, error) {
		atomic.AddInt32(&calls, 1)
		for i := 0; i < 1_000_000; i++ {
			_ = i
		}
		return len(key), nil
	}, 20*time.Millisecond)

	c.Get("abc") // 预热
	time.Sleep(40 * time.Millisecond)

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			c.Get("abc")
		}()
	}
	wg.Wait()

	// 第一次加载 + 过期后重新 singleflight 加载
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("loader=%d, want 2 (1 initial + 1 after expiry)", got)
	}
}

func TestExpiredCache_Refresh(t *testing.T) {
	callCount := 0
	c := NewExpiredCache[string, int](func(key string) (int, error) {
		callCount++
		return callCount, nil
	}, time.Hour)

	c.Get("k") // 1
	c.Refresh("k")
	if callCount != 2 {
		t.Fatalf("Refresh loader=%d, want 2", callCount)
	}
}

func TestExpiredCache_Operations(t *testing.T) {
	c := NewExpiredCache[int, string](nil, time.Hour)

	c.Set(1, "a")
	c.Set(2, "b")
	if c.Len() != 2 {
		t.Fatalf("Len=%d", c.Len())
	}

	c.Delete(1)
	if c.Contains(1) || c.Len() != 1 {
		t.Fatal("Delete failed")
	}

	c.Clear()
	if c.Len() != 0 {
		t.Fatal("Clear failed")
	}
}
