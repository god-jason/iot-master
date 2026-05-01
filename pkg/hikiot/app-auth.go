package hikvideo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (c *Client) GetAppAccessToken(ctx context.Context) (string, error) {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	if time.Now().Unix() < c.tokenExp-60 {
		return c.token, nil
	}

	body, _ := json.Marshal(map[string]string{
		"appKey":    c.AppKey,
		"appSecret": c.AppSecret,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST",
		c.BaseURL+"/auth/exchangeAppToken",
		bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res TokenResp
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	if res.Code != 0 {
		return "", fmt.Errorf("getAppAccessToken error code %d %s", res.Code, res.Msg)
	}

	c.token = res.Data.AppAccessToken
	c.tokenExp = time.Now().Unix() + res.Data.ExpiresIn*3600

	return c.token, nil
}
