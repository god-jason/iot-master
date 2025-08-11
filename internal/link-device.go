package internal

type LinkDevice struct {
	Id        string         `json:"id"`
	ProductId string         `json:"product_id"`
	Station   map[string]any `json:"station,omitempty"`
}
