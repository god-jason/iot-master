package hikvideo

import (
	"context"
	"encoding/json"
	"fmt"
)

type CaptureResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		CaptureUrl string `json:"captureUrl"`
		TraceId    string `json:"traceId"`
	} `json:"data"`
}

func (c *Client) Capture(ctx context.Context, sn string, ch int) (string, error) {

	resp, err := c.Request(ctx, "POST", "/device/direct/v1/captureImage/captureImage", nil, map[string]interface{}{
		"deviceSerial": sn,
		"payload": map[string]interface{}{
			"channelNo": ch,
		},
	})

	if err != nil {
		return "", err
	}

	var res CaptureResp
	err = json.Unmarshal(resp, &res)
	if err != nil {
		return "", err
	}

	if res.Code != 0 {
		return "", fmt.Errorf("capture error code %d %s", res.Code, res.Msg)
	}

	return res.Data.CaptureUrl, nil
}
