package event

import (
	"github.com/zgwit/iot-master/log"
	"github.com/zgwit/iot-master/pool"
)

// Emitter 事件监听器
type Emitter[T any] struct {
	id        int
	listeners map[int]func(t T)

	Sync bool //同步执行，默认不需要
}

// Emit 发送消息
func (e *Emitter[T]) Emit(t T) {
	if e.listeners != nil {
		for _, fn := range e.listeners {
			if e.Sync {
				fn(t)
				continue
			}

			if pool.Pool == nil {
				go fn(t)
				continue
			}

			//使用线程池执行
			err := pool.Insert(func() {
				fn(t)
			})

			//执行失败，就用原始办法执行
			if err != nil {
				log.Error(err)
				go fn(t)
			}
		}
	}
}

// On 监听消息
func (e *Emitter[T]) On(fn func(t T)) int {
	if e.listeners == nil {
		e.listeners = make(map[int]func(t T))
	}
	e.id++
	e.listeners[e.id] = fn
	return e.id
}

// Once 监听消息（仅一次）
func (e *Emitter[T]) Once(fn func(t T)) int {
	var id int
	id = e.On(func(t T) {
		fn(t)
		e.Off(id)
	})
	return id
}

// Off 删除监听
func (e *Emitter[T]) Off(id int) {
	if e.listeners != nil {
		delete(e.listeners, id)
	}
}

// Clear 取消所有监听
func (e *Emitter[T]) Clear() {
	e.listeners = nil
}
