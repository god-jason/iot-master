package protocol

type Request struct {
	MsgId string `json:"msg_id,omitempty"`
}

type Response struct {
	MsgId string `json:"msg_id"`
	Error string `json:"error,omitempty"`
}

type SyncRequest struct {
	MsgId    string `json:"msg_id,omitempty"`
	DeviceId string `json:"device_id"`
}

type SyncResponse struct {
	MsgId    string         `json:"msg_id"`
	Error    string         `json:"error,omitempty"`
	DeviceId string         `json:"device_id"`
	Values   map[string]any `json:"values"`
}

type ReadRequest struct {
	MsgId    string   `json:"msg_id,omitempty"`
	DeviceId string   `json:"device_id"`
	Points   []string `json:"points"`
}

type ReadResponse struct {
	MsgId    string         `json:"msg_id"`
	Error    string         `json:"error,omitempty"`
	DeviceId string         `json:"device_id"`
	Values   map[string]any `json:"values"`
}

type WriteRequest struct {
	MsgId    string         `json:"msg_id,omitempty"`
	DeviceId string         `json:"device_id"`
	Values   map[string]any `json:"values"`
}

type WriteResponse struct {
	MsgId    string          `json:"msg_id"`
	Error    string          `json:"error,omitempty"`
	DeviceId string          `json:"device_id"`
	Result   map[string]bool `json:"result"`
}

type ActionRequest struct {
	MsgId      string         `json:"msg_id,omitempty"`
	DeviceId   string         `json:"device_id"`
	Action     string         `json:"action"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type ActionResponse struct {
	MsgId    string         `json:"msg_id"`
	Error    string         `json:"error,omitempty"`
	DeviceId string         `json:"device_id"`
	Result   map[string]any `json:"result,omitempty"`
}
