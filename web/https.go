package web

import (
	"github.com/god-jason/iot-master/config"
	"github.com/god-jason/iot-master/log"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
)

func ServeTLS() error {
	cert := config.GetString(MODULE, "cert")
	key := config.GetString(MODULE, "key")

	log.Info("Web Server tls", cert, key)
	//return Engine.RunTLS(":443", cert, key)
	Server = &http.Server{Addr: ":https", Handler: Engine.Handler()}
	return Server.ListenAndServeTLS(cert, key)
}

func ServeLetsEncrypt() error {
	hosts := config.GetStringSlice(MODULE, "hosts")
	log.Info("Web Server with LetsEncrypt", hosts)

	//初始化autocert
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Email:      config.GetString(MODULE, "email"),
		HostPolicy: autocert.HostWhitelist(hosts...),
		Prompt:     autocert.AcceptTOS,
	}
	//return autotls.RunWithManager(Engine, manager)

	Server = &http.Server{
		Addr:      ":https",
		Handler:   Engine.Handler(),
		TLSConfig: manager.TLSConfig(),
	}

	return Server.ListenAndServeTLS("", "")
}
