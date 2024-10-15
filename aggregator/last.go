package aggregator

type last struct {
	value any
	dirty bool

	result any
}

func (a *last) Push(value any) error {
	a.value = value
	a.dirty = true
	return nil
}

func (a *last) Snap() {
	if !a.dirty {
		return
	}
	a.result = a.value
	a.dirty = false
}

func (a *last) Pop() any {
	return a.result
}
