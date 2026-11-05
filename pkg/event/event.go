package event

// EventBus 事件总线公共接口。
// eventBus（通配符 trie 版）和 simpleBus（精确匹配 map 版）都实现它。
// 外部包通过 NewEventBus() / NewSimple() 获取私有实现的指针，
// 持有此接口的变量即可调用统一的 Subscribe / Publish / Unsubscribe 方法。
type EventBus interface {
	Subscribe(filter string, handler Handler) func()
	Unsubscribe(filter string, handler Handler)
	Publish(topic string, payload any)
	Reset()
}
