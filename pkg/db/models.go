package db

var models []any

func Register(m ...any) {
	models = append(models, m...)
}
