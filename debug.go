package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/god-jason/iot-master/pkg/log"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	go func() {
		log.Fatal(http.ListenAndServe(":8088", mux))
	}()
}
