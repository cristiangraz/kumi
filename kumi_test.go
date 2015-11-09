package kumi

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
)

func TestKumi(t *testing.T) {
	k := New(&dummyRouter{})

	srv := &http.Server{
		Addr: ":8080",
		TLSConfig: &tls.Config{
			// Use at least one cipher suite so that http2.ConfigureServer
			// sets tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			},
			MinVersion: tls.VersionTLS12,
		},
	}
	srv2 := &http.Server{
		Addr:    ":8081",
		Handler: http.DefaultServeMux,
	}

	k.prep(srv, srv2)

	if !reflect.DeepEqual(http.DefaultServeMux, srv.Handler) {
		t.Errorf("TestKumi: Expected not passing a handler would set the handler default to http.DefaultServeMux")
	}

	hasCipherSuite := false
	for _, s := range srv.TLSConfig.CipherSuites {
		if s == tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 {
			hasCipherSuite = true
			break
		}
	}

	// TLS connections should be configured to use HTTP/2
	// tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 is a required cipher suite that should be added
	if !hasCipherSuite {
		t.Errorf("TestKumi: Expected http2.ConfigureServer to set %d cipher suite", tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
	}
}
