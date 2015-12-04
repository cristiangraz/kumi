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
}
