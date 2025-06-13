package link

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
)

func init() {
	api.Register("GET", "iot/link/list", curd.ApiList[Link]())
	api.Register("POST", "iot/link/search", curd.ApiSearch[Link]())
	api.Register("POST", "iot/link/create", curd.ApiCreate[Link]())
	api.Register("GET", "iot/link/:id", curd.ApiGet[Link]())
	api.Register("POST", "iot/link/:id", curd.ApiUpdate[Link]("id", "name", "description", "linker", "disabled", "protocol", "protocol_options"))
	api.Register("GET", "iot/link/:id/delete", curd.ApiDelete[Link]())
	api.Register("GET", "iot/link/:id/enable", curd.ApiDisable[Link](false))
	api.Register("GET", "iot/link/:id/disable", curd.ApiDisable[Link](true))

}
