package lib

import "sync"

type Pool[T any] struct {
	Init int
	Cap  int
	Idle int
	New  func() T
	
	lock sync.RWMutex
}

type poolItem[T any] struct {
	item T
}

func (p *Pool[T]) Obtain(fn func()) T {
	var t T
	return t
}
