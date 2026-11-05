package lib

import "sync"

type RingBuffer[T any] struct {
	buf   []T
	size  int
	head  int
	tail  int
	count int
	mu    sync.Mutex
}

func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{
		buf:  make([]T, size),
		size: size,
	}
}

func (rb *RingBuffer[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buf[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	} else {
		rb.head = (rb.head + 1) % rb.size
	}
}

func (rb *RingBuffer[T]) Pop() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		var zero T
		return zero, false
	}

	item := rb.buf[rb.head]
	rb.head = (rb.head + 1) % rb.size
	rb.count--

	return item, true
}

func (rb *RingBuffer[T]) Peek() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		var zero T
		return zero, false
	}

	return rb.buf[rb.head], true
}

func (rb *RingBuffer[T]) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

func (rb *RingBuffer[T]) Cap() int {
	return rb.size
}

func (rb *RingBuffer[T]) IsEmpty() bool {
	return rb.Len() == 0
}

func (rb *RingBuffer[T]) IsFull() bool {
	return rb.Len() == rb.size
}

func (rb *RingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.tail = 0
	rb.count = 0
}

func (rb *RingBuffer[T]) Values() []T {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	result := make([]T, 0, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.size
		result = append(result, rb.buf[idx])
	}
	return result
}
