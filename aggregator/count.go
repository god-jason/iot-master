package aggregator

type count struct {
	count  int64
	result int64
}

func (a *count) Push(value any) error {
	if value != nil {
		a.count++
	}
	return nil
}

func (a *count) Snap() {
	a.result = a.count
	a.count = 0
}

func (a *count) Pop() any {
	return a.result
}
