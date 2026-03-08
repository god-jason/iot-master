package lib

import (
	"sync"
	"time"
)

type cacheLoaderItem[T any] struct {
	value *T
	err   error

	expireAt int64
}

type CacheLoaderFunc[T any] func(key string) (*T, error)

type CacheLoader[T any] struct {
	items map[string]*cacheLoaderItem[T]
	lock  sync.Mutex

	Timeout int64
	Loader  CacheLoaderFunc[T]
}

func (c *CacheLoader[T]) Invalid(key string) {
	if item, ok := c.items[key]; ok {
		item.expireAt = time.Now().Unix()
	}
}

func (c *CacheLoader[T]) Load(key string) (*T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if item, ok := c.items[key]; ok {
		if time.Now().Unix() < item.expireAt {
			return item.value, item.err
		}
	}

	//忘记初始化了。。。
	if c.items == nil {
		c.items = make(map[string]*cacheLoaderItem[T])
	}

	item := &cacheLoaderItem[T]{}
	c.items[key] = item

	//正式加载
	item.value, item.err = c.Loader(key)
	item.expireAt = time.Now().Unix() + c.Timeout

	return item.value, item.err
}
