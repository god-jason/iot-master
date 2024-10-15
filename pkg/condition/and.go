package condition

import "github.com/zgwit/iot-master/pkg/exception"

type And struct {
	Compares []*Compare `json:"compares,omitempty"`
}

func (a *And) Init() error {
	for _, c := range a.Compares {
		err := c.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *And) Eval(ctx map[string]any) (bool, error) {
	if len(a.Compares) == 0 {
		return false, exception.New("没有对比")
	}
	for _, c := range a.Compares {
		ret, err := c.Eval(ctx)
		if err != nil {
			return ret, err
		}
		if !ret {
			return false, nil
		}
	}
	return true, nil
}
