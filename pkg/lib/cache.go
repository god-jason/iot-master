package lib

import (
	"sync"
	"time"
)

type cacheItem[T any] struct {
	value    *T
	expireAt int64
}

type Cache[T any] struct {
	mp sync.Map
	//tm *time.Timer

	Timeout int64
}

func (c *Cache[T]) Delete(key string) {
	c.mp.Delete(key)
}

func (c *Cache[T]) Load(key string) (*T, bool) {
	if i, ok := c.mp.Load(key); ok {
		item := i.(*cacheItem[T])
		if time.Now().Unix() > item.expireAt {
			c.mp.Delete(key)
			return nil, false
		}
		return item.value, true
	}
	return nil, false
}

func (c *Cache[T]) Store(key string, value *T) {
	c.mp.Store(key, &cacheItem[T]{
		value:    value,
		expireAt: time.Now().Unix() + c.Timeout,
	})
}

//func (c *Cache[T]) checkExpire() {
//	hasRemain := false
//
//	now := time.Now().Unix()
//	c.mp.Range(func(key, value any) bool {
//		item := value.(*cacheItem[T])
//		if now >= item.expireAt {
//			c.mp.Delete(item)
//		} else {
//			hasRemain = true
//		}
//		return true
//	})
//
//	if hasRemain {
//		c.tm = time.AfterFunc(time.Second, func() {
//			c.checkExpire()
//		})
//	} else {
//		c.tm = nil
//	}
//}
