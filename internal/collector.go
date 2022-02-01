package interval

type Collector struct {
	Disabled bool   `json:"disabled"`
	Type     string `json:"type"` //interval, clock, crontab
	Interval int    `json:"interval"`
	Clock    int    `json:"clock"`
	Crontab  string `json:"crontab"`

	Code    int `json:"code"`
	Address int `json:"address"`
	//TODO Address2
	Length  int `json:"length"`

	//TODO Filters

}

func (c *Collector) Start() error {
	return nil
}

func (c *Collector) Stop() error {
	return nil
}