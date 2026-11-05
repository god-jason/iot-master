package db

var models []any

// Register 注册 model 给 Startup 时自动 Sync（sync=true 时生效）
func Register(m ...any) {
	models = append(models, m...)
}
