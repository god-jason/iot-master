package web

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/logger"

	"golang.org/x/crypto/acme/autocert"
)

// Run 启动HTTP服务器
func (e *Engine) Run(addr string) error {
	logger.Info("服务启动中", "address", addr)
	return e.Engine.Run(addr)
}

// RunTLS 启动HTTPS服务器
func (e *Engine) RunTLS(addr, certFile, keyFile string) error {
	logger.Info("服务启动中(TLS)", "address", addr)
	return e.Engine.RunTLS(addr, certFile, keyFile)
}

// RunWithACME 启动支持 Let's Encrypt 的服务器
// 同时监听 80 端口用于 ACME HTTP-01 验证和 HTTP→HTTPS 跳转
func (e *Engine) RunWithACME() error {
	addr := fmt.Sprintf("%s:%d", config.GetString(MODULE, "host"), config.GetInt(MODULE, "port"))
	domains := config.GetStringSlice(MODULE, "domains")
	email := config.GetString(MODULE, "email")

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(domains...),
		Email:      email,
	}

	httpServer := &http.Server{
		Addr:    ":80",
		Handler: m.HTTPHandler(nil),
	}
	go func() {
		logger.Info("HTTP服务启动中(ACME验证)", "address", ":80")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP服务启动失败", "error", err)
		}
	}()

	tlsConfig := &tls.Config{GetCertificate: m.GetCertificate}
	httpsServer := &http.Server{
		Addr:      addr,
		Handler:   e.Engine,
		TLSConfig: tlsConfig,
	}

	logger.Info("服务启动中(ACME)", "address", addr, "domains", domains, "email", email)
	return httpsServer.ListenAndServeTLS("", "")
}

// RunServer 根据 viper 配置启动服务器
func (e *Engine) RunServer() error {
	addr := fmt.Sprintf("%s:%d", config.GetString(MODULE, "host"), config.GetInt(MODULE, "port"))

	switch {
	case config.GetBool(MODULE, "lets_encrypt"):
		return e.RunWithACME()
	case config.GetBool(MODULE, "https") && config.GetString(MODULE, "cert_file") != "" && config.GetString(MODULE, "key_file") != "":
		return e.RunTLS(addr, config.GetString(MODULE, "cert_file"), config.GetString(MODULE, "key_file"))
	default:
		return e.Run(addr)
	}
}
