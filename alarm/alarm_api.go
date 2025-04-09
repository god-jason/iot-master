package alarm

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
)

func init() {
	api.Register("GET", "iot/alarm/list", curd.ApiList[Alarm]())
	api.Register("POST", "iot/alarm/search", curd.ApiSearch[Alarm]())
	api.Register("GET", "iot/alarm/:id/delete", curd.ApiDelete[Alarm]())
}
