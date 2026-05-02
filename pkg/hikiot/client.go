package hikvideo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	Encrypted bool

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

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (c *Client) CheckAppAccessToken(ctx context.Context) error {
	if c.token != "" {
		if time.Now().Unix() < c.tokenExp-60 {
			return nil
		}
	}

	_, err := c.GetAppAccessToken(ctx)
	return err
}

func (c *Client) CheckUserAccessToken(ctx context.Context) error {
	if c.userToken != "" {
		if time.Now().Unix() < c.userTokenExp-60 {
			return nil
		}
	}

	code, err := c.GetAuthCode(ctx)
	if err != nil {
		return err
	}

	_, err = c.GetUserAccessToken(ctx, code)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Request(ctx context.Context, method string, uri string, query url.Values, body map[string]any) (map[string]any, error) {
	var err error

	//URL参数
	_query := ""
	if len(query) > 0 {
		_query = query.Encode()
		if c.Encrypted {
			_query, err = EncryptByPrivateKey(_query, c.AppSecret)
			if err != nil {
				return nil, err
			}
			_query = "querySecret=" + url.QueryEscape(_query)
		}
	}

	u := c.BaseURL + uri
	if _query != "" {
		u += "?" + _query
	}

	//转换body
	_body := ""
	if len(body) > 0 {
		b, e := json.Marshal(body)
		if e != nil {
			return nil, e
		}
		_body = string(b)

		if c.Encrypted {
			_body, err = EncryptByPrivateKey(_body, c.AppSecret)
			if err != nil {
				return nil, err
			}

			b, e = json.Marshal(map[string]any{
				"bodySecret": _body,
			})
			if e != nil {
				return nil, e
			}
			_body = string(b)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hikiot http status %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res Response
	err = json.Unmarshal(buf, &res)
	if err != nil {
		return nil, err
	}

	if res.Code != 0 {
		return nil, fmt.Errorf("hikiot response %d %s", res.Code, res.Msg)
	}

	switch data := res.Data.(type) {
	case map[string]interface{}:
		return data, nil
	case string:
		str, err := DecryptByPrivateKey(data, c.AppSecret)
		if err != nil {
			return nil, err
		}
		var vs map[string]interface{}
		err = json.Unmarshal([]byte(str), &vs)
		return vs, err
	default:
		return nil, fmt.Errorf("hikiot response %d %s", res.Code, res.Msg)
	}
}
