package lib

import "sync/atomic"

type RingBuffer[T any] struct {
	Size   int32
	Reader int32
	Writer int32
	Buffer []T
}

func NewRingBuffer[T any](size int) *RingBuffer[T] {
	rb := &RingBuffer[T]{
		Buffer: make([]T, size),
	}
	return rb
}

func (r *RingBuffer[T]) Write(value T) {
	current := atomic.LoadInt32(&r.Writer)
	r.Buffer[current] = value
	next := (current + 1) % r.Size
	atomic.StoreInt32(&r.Writer, next)
}

func (r *RingBuffer[T]) seekReader(delta int32) {
	current := atomic.LoadInt32(&r.Reader)
	expected := (current + delta) % r.Size
	atomic.StoreInt32(&r.Reader, expected)
}

func (r *RingBuffer[T]) Read() T {
	defer r.seekReader(1)
	return r.Buffer[atomic.LoadInt32(&r.Reader)]
}

func (r *RingBuffer[T]) Latest() T {
	return r.Buffer[(atomic.LoadInt32(&r.Writer)-1)%r.Size]
}

func (r *RingBuffer[T]) Oldest() T {
	return r.Buffer[atomic.LoadInt32(&r.Writer)]
}

func (r *RingBuffer[T]) Overwrite(v T) {
	r.Buffer[atomic.LoadInt32(&r.Writer)] = v
}
