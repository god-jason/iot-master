package cache

// ErrLoaderNotSet loader 未设置错误
var ErrLoaderNotSet = &CacheError{"loader not set"}

// CacheError 缓存错误
type CacheError struct {
	message string
}

func (e *CacheError) Error() string {
	return e.message
}
