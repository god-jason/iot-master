package lib

import "sync"

type Map[T any] struct {
	container map[string]*T
	lock      sync.RWMutex
}

func (c *Map[T]) Load(name string) *T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.container == nil {
		return nil
	}
	return c.container[name]
}

func (c *Map[T]) Store(name string, value *T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.container == nil {
		c.container = make(map[string]*T)
	}
	c.container[name] = value
}

func (c *Map[T]) Range(iterator func(name string, item *T) bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.container == nil {
		return
	}
	for k, v := range c.container {
		if !iterator(k, v) {
			break
		}
	}
}

func (c *Map[T]) Delete(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.container == nil {
		return
	}
	delete(c.container, name)
}

func (c *Map[T]) DeleteDirectly(name string) {
	if c.container == nil {
		return
	}
	delete(c.container, name)
}

func (c *Map[T]) LoadAndStore(name string, value *T) *T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.container == nil {
		c.container = make(map[string]*T)
	}
	ret := c.container[name]
	c.container[name] = value
	return ret
}

func (c *Map[T]) LoadAndDelete(name string) *T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.container == nil {
		return nil
	}
	ret := c.container[name]
	delete(c.container, name)
	return ret
}

func (c *Map[T]) Len() int {
	return len(c.container)
}

func (c *Map[T]) Map() map[string]*T {
	return c.container
}

func (c *Map[T]) Clear() {
	c.lock.RLock()
	defer c.lock.RUnlock()
	c.container = nil
}
