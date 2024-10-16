package types

type Addr interface {
}

type Driver interface {
	Parse(addr string) (Addr, error)
	Fetch(station map[string]any, addr string, size int) (values map[string]any, err error)
	Write(station map[string]any, addr string, value any) (err error)
}
