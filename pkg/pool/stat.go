package pool

type Info struct {
	Cap     int `json:"cap,omitempty"`
	Free    int `json:"free,omitempty"`
	Running int `json:"running,omitempty"`
	Waiting int `json:"waiting,omitempty"`
}

func Stats() *Info {
	return &Info{
		Cap:     Pool.Cap(),
		Free:    Pool.Free(),
		Running: Pool.Running(),
		Waiting: Pool.Waiting(),
	}
}
