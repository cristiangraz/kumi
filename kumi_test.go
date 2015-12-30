package kumi

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
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

func TestBodylessResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	writer := BodylessResponseWriter{w}

	written, err := writer.Write([]byte("hello"))
	if err != nil {
		t.Errorf("TestBodylessResponseWriter: Didn't expect an error. Error: %s", err)
	}

	if written != 5 {
		t.Errorf("TestBodylessResponseWriter: Expected 5 bytes to be recorded. Written: %d", written)
	}

	if len(w.Body.Bytes()) > 0 {
		t.Error("TestBodylessResponseWriter: Didn't expect any bytes to be written")
	}
}
