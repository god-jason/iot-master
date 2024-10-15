package lib

import "sync"

type Values struct {
	data map[string]any
	lock sync.RWMutex
}

func (p *Values) Put(id string, val any) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.data == nil {
		p.data = make(map[string]any)
	}
	p.data[id] = val
}

func (p *Values) Get(id string) any {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.data[id]
}

func (p *Values) GetAll() any {
	return p.data
}

func (p *Values) Merge(props map[string]any) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.data == nil {
		p.data = make(map[string]any)
	}

	for k, v := range props {
		p.data[k] = v
	}
}

func (p *Values) Clear() {
	p.data = make(map[string]any)
}
