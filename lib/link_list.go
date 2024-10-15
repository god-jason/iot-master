package lib

import (
	"fmt"
	"sync"
)

type linkNode[T any] struct {
	prev, next *linkNode[T]
	value      T
}

type LinkList[T any] struct {
	head, tail *linkNode[T]
	size       int
	lock       sync.RWMutex
}

func (l *LinkList[T]) Add(v T, num int) {
	// when head is null
	if l.size == num {
		l.Push(v)
		return
	}

	if l.size < num {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	node := linkNode[T]{value: v, prev: nil, next: nil}
	l.size++

	if num == 0 {
		node.next = l.head
		l.head.prev = &node
		l.head = &node
		return
	}

	next := l.head
	for i := 1; i < num; i++ {
		next = next.next
	}

	node.prev = next
	node.next = next.next
	next.next.prev = &node
	next.next = &node

	if node.next == nil {
		l.tail = &node
	}
}

func (l *LinkList[T]) Remove(index int) {

	if index == 0 {
		l.Pop()
		return
	}

	if index == l.size {
		l.Dequeue()
		return
	}

	if l.size < index {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	ptr := l.head
	for i := 1; i < index; i++ {
		ptr = ptr.next
	}

	ptr.next.prev = ptr.prev
	ptr.prev.next = ptr.next
	l.size--
}

func (l *LinkList[T]) Push(v T) {

	l.lock.Lock()
	defer l.lock.Unlock()

	node := &linkNode[T]{value: v}
	if l.head == nil {
		l.head = node
		l.tail = l.head
	} else {
		node.prev = l.tail
		l.tail.next = node
		l.tail = node
	}
	l.size++
}

func (l *LinkList[T]) Pop() {
	if l.tail == nil {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	if l.tail.prev == nil {
		l.tail = nil
		l.head = nil
	} else {
		l.tail = l.tail.prev
		l.tail.next = nil
	}
	l.size--
}

func (l *LinkList[T]) Enqueue(v T) {
	l.Push(v)
}

func (l *LinkList[T]) Dequeue() {
	if l.head == nil {
		return
	}

	l.lock.Lock()
	defer l.lock.Unlock()

	l.head = l.head.next
	l.size--
}

func (l *LinkList[T]) Size() int {
	return l.size
}

func (l *LinkList[T]) Get(index int) T {
	if l.size <= index {
		panic(fmt.Errorf("链表越界访问 index:%d size:%d", index, l.size))
	}

	l.lock.RLock()
	defer l.lock.RUnlock()

	ptr := l.head
	for i := 1; i <= index; i++ {
		ptr = ptr.next
	}

	return ptr.value
}

func (l *LinkList[T]) GetAll() []T {
	if l.size == 0 {
		return make([]T, 0)
	}

	l.lock.RLock()
	defer l.lock.RUnlock()

	items := make([]T, l.size)
	node := l.head

	for i := 0; i <= l.size; i++ {
		items[i] = node.value

		if node.next == nil {
			break
		}

		node = (*node).next
	}

	return items
}

func (l *LinkList[T]) Walk(fn func(v T) bool) {
	if l.size == 0 {
		return
	}

	l.lock.RLock()
	defer l.lock.RUnlock()

	node := l.head

	for i := 0; i <= l.size; i++ {
		if !fn(node.value) {
			break
		}
		if node.next == nil {
			break
		}
		node = node.next
	}
}
