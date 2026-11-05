package cache

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCache_BasicGetSet(t *testing.T) {
	c := NewCache[string, int](func(key string) (int, error) {
		return len(key), nil
	})

	// 首次 miss → loader
	v, err := c.Get("abc")
	if err != nil || v != 3 {
		t.Fatalf("miss: got %d %v, want 3 <nil>", v, err)
	}

	// 再 Get → 命中缓存，loader 不应被调用
	v, err = c.Get("abc")
	if err != nil || v != 3 {
		t.Fatalf("hit: got %d %v, want 3 <nil>", v, err)
	}

	// 手动 Set 覆盖
	c.Set("abc", 42)
	v, err = c.Get("abc")
	if err != nil || v != 42 {
		t.Fatalf("after Set: got %d %v, want 42 <nil>", v, err)
	}
}

func TestCache_LoaderNotSet(t *testing.T) {
	c := NewCache[string, int](nil)
	_, err := c.Get("x")
	if !errors.Is(err, ErrLoaderNotSet) {
		t.Fatalf("want ErrLoaderNotSet, got %v", err)
	}

	// SetLoader 后置注入后应正常工作
	c.SetLoader(func(key string) (int, error) { return 7, nil })
	v, err := c.Get("x")
	if err != nil || v != 7 {
		t.Fatalf("after SetLoader: got %d %v, want 7 <nil>", v, err)
	}
}

func TestCache_ErrorCaching(t *testing.T) {
	fail := errors.New("boom")
	callCount := 0
	c := NewCache[string, int](func(key string) (int, error) {
		callCount++
		return 0, fail
	})

	// 第一次 loader 失败
	_, err := c.Get("k")
	if !errors.Is(err, fail) {
		t.Fatalf("want boom, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, want 1", callCount)
	}

	// 后续 Get 应直接返回缓存的错误，loader 不被调用
	_, err = c.Get("k")
	_, err = c.Get("k")
	if !errors.Is(err, fail) {
		t.Fatalf("want boom from cache, got %v", err)
	}
	if callCount != 1 {
		t.Fatalf("loader should NOT be retried, got %d calls", callCount)
	}

	// Delete 后重新加载
	c.Delete("k")
	_, _ = c.Get("k")
	if callCount != 2 {
		t.Fatalf("after Delete: loader called %d times, want 2", callCount)
	}
}

func TestCache_Refresh(t *testing.T) {
	callCount := 0
	c := NewCache[string, int](func(key string) (int, error) {
		callCount++
		return callCount, nil
	})

	_, _ = c.Get("k") // 1
	_, _ = c.Get("k") // 命中
	v, err := c.Refresh("k")
	if err != nil || v != 2 {
		t.Fatalf("Refresh: got %d %v, want 2 <nil>", v, err)
	}
	if callCount != 2 {
		t.Fatalf("loader calls=%d, want 2", callCount)
	}
}

func TestCache_SetErr(t *testing.T) {
	myErr := errors.New("manually set")
	c := NewCache[string, int](nil)
	c.SetErr("k", myErr)

	if !c.Contains("k") {
		t.Fatal("Contains should be true for error-marked entry")
	}
	_, err := c.Get("k")
	if !errors.Is(err, myErr) {
		t.Fatalf("want manually set err, got %v", err)
	}
}

func TestCache_Operations(t *testing.T) {
	c := NewCache[int, string](func(key int) (string, error) {
		return "", nil
	})

	// Set / Contains / Len / Keys
	c.Set(1, "a")
	c.Set(2, "b")

	if !c.Contains(1) || !c.Contains(2) || c.Contains(3) {
		t.Fatal("Contains mismatch")
	}
	if c.Len() != 2 {
		t.Fatalf("Len=%d, want 2", c.Len())
	}
	if len(c.Keys()) != 2 {
		t.Fatalf("Keys len=%d, want 2", len(c.Keys()))
	}

	// Delete
	c.Delete(1)
	if c.Contains(1) || c.Len() != 1 {
		t.Fatal("after Delete")
	}

	// Clear
	c.Clear()
	if c.Len() != 0 {
		t.Fatal("after Clear")
	}
}

func TestCache_Singleflight(t *testing.T) {
	var calls int32
	c := NewCache[int, int](func(key int) (int, error) {
		atomic.AddInt32(&calls, 1)
		// 故意延迟，让并发请求真的撞在一起
		for i := 0; i < 1_000_000; i++ {
			_ = i
		}
		return key * 2, nil
	})

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			v, err := c.Get(7)
			if err != nil || v != 14 {
				t.Errorf("got %d %v, want 14 <nil>", v, err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("singleflight broken: loader called %d times, want 1", got)
	}
}

func TestCache_SingleflightError(t *testing.T) {
	fail := errors.New("db down")
	var calls int32
	c := NewCache[int, int](func(key int) (int, error) {
		atomic.AddInt32(&calls, 1)
		// 延迟确保 followers 与 inflight 碰撞
		for i := 0; i < 1_000_000; i++ {
			_ = i
		}
		return 0, fail
	})

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, _ = c.Get(7)
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("error + singleflight: loader called %d times, want 1", got)
	}
}
