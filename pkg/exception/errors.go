package exception

import (
	"errors"
	"runtime/debug"
)

var _release = false

type stackError struct {
	err error
	//message string
	stack []byte
}

func (e *stackError) Error() string {
	return e.err.Error() + "\n" + string(e.stack)
}

func (e *stackError) Unwrap() error {
	return e.err
}

func New(message string) error {
	if _release {
		return errors.New(message)
	}
	return &stackError{
		err:   errors.New(message),
		stack: debug.Stack(),
	}
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}
	if _release {
		return err
	}
	var e *stackError
	if errors.As(err, &e) { //避免多次封装
		return e
	}
	return &stackError{
		err:   err,
		stack: debug.Stack(),
	}
}

func SetRelease(r bool) {
	_release = r
}
