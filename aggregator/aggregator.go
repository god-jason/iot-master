package aggregator

import "github.com/god-jason/iot-master/pkg/exception"

type Aggregator interface {
	Push(any) error
	Pop() any
	Snap() //快照
}

// New 新建
func New(typ string) (agg Aggregator, err error) {
	switch typ {
	case "inc", "increase":
		agg = &inc{}
	case "dec", "decrease":
		agg = &dec{}
	case "sum", "acc":
		agg = &sum{}
	case "avg", "average":
		agg = &avg{}
	case "cnt", "count":
		agg = &count{}
	case "min", "minimum":
		agg = &minimum{}
	case "max", "maximum":
		agg = &maximum{}
	case "first":
		agg = &first{}
	case "last":
		agg = &last{}
	default:
		err = exception.New("未知的聚合类型" + typ)
	}
	return
}
