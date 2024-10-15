package web

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func CreateReverseProxy(typ, address string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	//创建反向代理服务
	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			//req.URL.Path = u.Path
			//设置User-Agent
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
	}

	//支持unix socket加速
	if typ == "unix" {
		rp.Transport = &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", address)
		}}
	}

	return rp, nil
}
