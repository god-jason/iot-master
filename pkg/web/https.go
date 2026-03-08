package web

import (
	"net/http"

	"github.com/busy-cloud/boat/config"
	"github.com/busy-cloud/boat/log"
	"golang.org/x/crypto/acme/autocert"
)

func ServeTLS() error {
	cert := config.GetString(MODULE, "tls_cert")
	key := config.GetString(MODULE, "tls_key")

	log.Info("Web Server tls", cert, key)
	//return engine.RunTLS(":443", cert, key)
	Server = &http.Server{Addr: ":https", Handler: engine.Handler()}
	return Server.ListenAndServeTLS(cert, key)
}

func ServeAutoCert() error {
	hosts := config.GetStringSlice(MODULE, "hosts")
	log.Info("Web Server with LetsEncrypt", hosts)

	//初始化autocert
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Email:      config.GetString(MODULE, "email"),
		HostPolicy: autocert.HostWhitelist(hosts...),
		Prompt:     autocert.AcceptTOS,
	}
	//return autotls.RunWithManager(engine, manager)

	Server = &http.Server{
		Addr:      ":https",
		Handler:   engine.Handler(),
		TLSConfig: manager.TLSConfig(),
	}

	return Server.ListenAndServeTLS("", "")
}
