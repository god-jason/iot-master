package hikvideo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Client struct {
	BaseURL   string
	AppKey    string
	AppSecret string
	Username  string
	Password  string

	httpClient *http.Client

	token     string
	tokenExp  int64
	tokenLock sync.Mutex

	authCode string

	userToken    string
	userTokenExp int64
	userLock     sync.Mutex

	streamCache map[string]*streamItem
	cacheLock   sync.RWMutex
}

func NewClient(appKey, appSecret, username, password string) *Client {
	return &Client{
		BaseURL:   "https://open-api.hikiot.com",
		AppKey:    appKey,
		AppSecret: appSecret,
		Username:  username,
		Password:  password,

		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		streamCache: make(map[string]*streamItem),
	}
}

func (c *Client) Request(ctx context.Context, method string, uri string, query url.Values, body map[string]any) ([]byte, error) {

	if c.token == "" {
		_, err := c.GetAppAccessToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	if c.userToken == "" {
		code, err := c.GetAuthCode(ctx)
		if err != nil {
			return nil, err
		}
		_, err = c.GetUserAccessToken(ctx, code)
		if err != nil {
			return nil, err
		}
	}

	var err error

	//URL参数
	_query := ""
	if len(query) > 0 {
		_query = query.Encode()
		_query, err = EncryptByPrivateKey(_query, c.AppSecret)
		if err != nil {
			return nil, err
		}
	}

	u := c.BaseURL + uri
	if _query != "" {
		u += "?querySecret=" + _query
	}

	//转换body
	_body := ""
	if len(body) > 0 {
		b, e := json.Marshal(body)
		if e != nil {
			return nil, e
		}
		_body = string(b)
		_body, err = EncryptByPrivateKey(_body, c.AppSecret)
		if err != nil {
			return nil, err
		}
	}

	req, _ := http.NewRequestWithContext(ctx, method, u, bytes.NewReader([]byte(_body)))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("App-Access-Token", c.token)
	req.Header.Set("User-Access-Token", c.userToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	str, err := DecryptByPrivateKey(string(buf), c.AppSecret)
	if err != nil {
		return nil, err
	}

	return []byte(str), nil
}
