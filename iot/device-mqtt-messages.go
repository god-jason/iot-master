package iot

type Request struct {
	MsgId string `json:"msg_id,omitempty"`
}

type Response struct {
	MsgId string `json:"msg_id"`
	Error string `json:"error,omitempty"`
}

type SyncRequest struct {
	MsgId    string `json:"msg_id,omitempty"`
	DeviceId string `json:"device_id,omitempty"`
}

type SyncResponse struct {
	MsgId    string         `json:"msg_id"`
	Error    string         `json:"error,omitempty"`
	DeviceId string         `json:"device_id,omitempty"`
	Values   map[string]any `json:"values"`
}

type ReadRequest struct {
	MsgId    string   `json:"msg_id,omitempty"`
	DeviceId string   `json:"device_id,omitempty"`
	Points   []string `json:"points"`
}

type ReadResponse struct {
	MsgId    string         `json:"msg_id"`
	Error    string         `json:"error,omitempty"`
	DeviceId string         `json:"device_id,omitempty"`
	Values   map[string]any `json:"values"`
}

type WriteRequest struct {
	MsgId    string         `json:"msg_id,omitempty"`
	DeviceId string         `json:"device_id,omitempty"`
	Values   map[string]any `json:"values"`
}

type WriteResponse struct {
	MsgId    string          `json:"msg_id"`
	Error    string          `json:"error,omitempty"`
	DeviceId string          `json:"device_id,omitempty"`
	Result   map[string]bool `json:"result"`
}

type ActionRequest struct {
	MsgId      string         `json:"msg_id,omitempty"`
	DeviceId   string         `json:"device_id,omitempty"`
	Action     string         `json:"action"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type ActionResponse struct {
	MsgId    string `json:"msg_id"`
	Error    string `json:"error,omitempty"`
	DeviceId string `json:"device_id,omitempty"`
	Action   string `json:"action"`
	Result   any    `json:"result,omitempty"`
}

type SettingRequest struct {
	MsgId   string `json:"msg_id,omitempty"`
	Name    string `json:"name"`
	Content any    `json:"content"`
	Version int    `json:"version"`
}
type SettingResponse struct {
	MsgId   string `json:"msg_id"`
	Error   string `json:"error,omitempty"`
	Name    string `json:"name"`
	Content any    `json:"content"`
	Version int    `json:"version"`
}
