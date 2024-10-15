package lib

import (
	"sync"
	"time"
)

type ExpireCache struct {
	mp sync.Map
	tm *time.Timer

	Timeout int64
}

type cacheItem struct {
	value    any
	expireAt int64
}

func (c *ExpireCache) Delete(key string) {
	c.mp.Delete(key)
}

func (c *ExpireCache) Load(key string) (any, bool) {
	if i, ok := c.mp.Load(key); ok {
		item := i.(*cacheItem)
		item.expireAt = time.Now().Unix() + c.Timeout
		return item.value, true
	}
	return nil, false
}

func (c *ExpireCache) Store(key string, value any) {
	c.mp.Store(key, &cacheItem{value: value, expireAt: time.Now().Unix() + c.Timeout})

	if c.tm == nil {
		c.tm = time.AfterFunc(time.Second, func() {
			c.checkExpire()
		})
	}
}

func (c *ExpireCache) checkExpire() {
	hasRemain := false

	now := time.Now().Unix()
	c.mp.Range(func(key, value any) bool {
		item := value.(*cacheItem)
		if now >= item.expireAt {
			c.mp.Delete(item)
		} else {
			hasRemain = true
		}
		return true
	})

	if hasRemain {
		c.tm = time.AfterFunc(time.Second, func() {
			c.checkExpire()
		})
	} else {
		c.tm = nil
	}
}
