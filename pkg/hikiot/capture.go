package hikvideo

import (
	"context"
)

func (c *Client) Capture(ctx context.Context, sn string, ch int) (string, error) {
	err := c.CheckAppAccessToken(ctx)
	if err != nil {
		return "", err
	}
	err = c.CheckUserAccessToken(ctx)
	if err != nil {
		return "", err
	}

	resp, err := c.Request(ctx, "POST", "/device/direct/v1/captureImage/captureImage", nil, map[string]interface{}{
		"deviceSerial": sn,
		"payload": map[string]interface{}{
			"channelNo": ch,
		},
	})

	if err != nil {
		return "", err
	}

	return resp["captureUrl"].(string), nil
}
