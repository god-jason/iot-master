package aggregator

type first struct {
	value any
	dirty bool

	result any
}

func (a *first) Push(value any) error {
	if !a.dirty {
		return nil
	}

	a.value = value
	a.dirty = true
	return nil
}

func (a *first) Snap() {
	if !a.dirty {
		return
	}
	a.result = a.value
	a.dirty = false
}

func (a *first) Pop() any {
	return a.result
}
