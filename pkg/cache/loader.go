package cache

//LoaderFunc 缓存未命中时的加载回调函数
type LoaderFunc[K comparable, V any] func(key K) (V, error)
