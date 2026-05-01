package hikvideo

import "time"

type streamItem struct {
	URL    string
	Expire int64
}

func (c *Client) getCache(key string) (string, bool) {
	c.cacheLock.RLock()
	defer c.cacheLock.RUnlock()

	item, ok := c.streamCache[key]
	if !ok {
		return "", false
	}

	if time.Now().Unix() > item.Expire {
		return "", false
	}

	return item.URL, true
}

func (c *Client) setCache(key, url string) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	c.streamCache[key] = &streamItem{
		URL:    url,
		Expire: time.Now().Unix() + 300, // 默认5分钟
	}
}
