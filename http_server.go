package main

import (
	"fmt"
	"github.com/packethost/hegel/gxff"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sebest/xff"
	"net/http"
)

func ServeHTTP() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/_packet/healthcheck", healthCheckHandler)
	mux.HandleFunc("/_packet/version", versionHandler)
	mux.HandleFunc("/metadata", getMetadata)

	var handler http.Handler
	if len(gxff.TrustedProxies) > 0 {
		xffmw, _ := xff.New(xff.Options{
			AllowedSubnets: gxff.TrustedProxies,
		})

		handler = xffmw.Handler(mux)
	} else {
		handler = mux
	}
	http.Handle("/", handler)

	logger.With("port", *metricsPort).Info("Starting http server")
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", *metricsPort), nil)
		if err != nil {
			logger.Error(err, "failed to serve http")
			panic(err)
		}
	}()
}
