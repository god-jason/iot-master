package javascript

import (
	"github.com/busy-cloud/boat/exception"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/url"
)

var req require.Registry

func Runtime() *goja.Runtime {
	runtime := goja.New()

	//支持require
	req.Enable(runtime)

	//支持console.log
	console.Enable(runtime)

	//支持url解析
	url.Enable(runtime)

	return runtime
}

func Compile(src string) (*goja.Program, error) {
	return goja.Compile("", src, false)
}

func Run(p *goja.Program) (any, error) {
	ret, err := Runtime().RunProgram(p)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	return ret.Export(), nil
}

func Exec(src string) (any, error) {
	ret, err := Runtime().RunString(src)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	return ret.Export(), nil
}
