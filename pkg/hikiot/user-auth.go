package hikvideo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type AuthResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AppKey      string `json:"appKey"`
		RedirectUrl string `json:"redirectUrl"`
		AuthCode    string `json:"authCode"`
	} `json:"data"`
}

func (c *Client) GetAuthCode(ctx context.Context) (string, error) {
	c.userLock.Lock()
	defer c.userLock.Unlock()

	body, _ := json.Marshal(map[string]string{
		"appKey":      c.AppKey,
		"userName":    c.Username,
		"password":    c.Password,
		"redirectUrl": "https://open.hikiot.com/util",
	})

	req, _ := http.NewRequestWithContext(ctx, "POST",
		c.BaseURL+"/auth/third/applyAuthCode",
		bytes.NewReader(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Access-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res AuthResp
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return "", err
	}

	if res.Code != 0 {
		return "", fmt.Errorf("GetAuthCode error code %d %s", res.Code, res.Msg)
	}

	c.authCode = res.Data.AuthCode

	return c.authCode, nil
}

type UserResp struct {
	AppKey           string `json:"appKey"`
	UserAccessToken  string `json:"userAccessToken"`
	RefreshUserToken string `json:"refreshUserToken"`
	ExpiresIn        int64  `json:"expiresIn"`
	TeamNo           string `json:"teamNo"`
	PersonNo         string `json:"personNo"`
	AccountNo        string `json:"accountNo"`
}

func (c *Client) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	c.userLock.Lock()
	defer c.userLock.Unlock()

	us := make(url.Values)
	us.Set("code", code)

	resp, err := c.Request(ctx, "GET", "/auth/third/code2Token", us, nil)
	if err != nil {
		return "", err
	}

	c.userToken = resp["userAccessToken"].(string)
	c.userTokenExp = time.Now().Unix() + resp["ExpiresIn"].(int64)*3600*12

	return c.userToken, nil
}
