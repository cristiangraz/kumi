package kumi

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestKumi(t *testing.T) {
	k := New(&testRouter{})

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

func TestNewContextHTTP(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/not-found-path", nil)
	req.Host = "exAmPlE.com"

	k := New(&testRouter{})
	c := k.NewContext(rec, req)

	if c.Request.Host != "example.com" {
		t.Errorf("TestNewContextHTTP: Expected Host to be lowercased. Given: %s", c.Request.Host)
	}

	if c.Request.URL.Host != "example.com" {
		t.Errorf("TestNewContextHTTP: Expected Host to be lowercased. Given: %s", c.Request.Host)
	}

	if c.Request.URL.Scheme != "http" {
		t.Errorf("TestNewContextHTTP: Expected scheme to be http. Given %s", c.Request.URL.Scheme)
	}

	k.ReturnContext(c)
}

func TestNewContextHTTPS(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/not-found-path", nil)
	req.Host = "Example.com"
	req.TLS = &tls.ConnectionState{}

	k := New(&testRouter{})
	c := k.NewContext(rec, req)

	if c.Request.Host != "example.com" {
		t.Errorf("TestNewContextHTTP: Expected Host to be lowercased. Given: %s", c.Request.Host)
	}

	if c.Request.URL.Host != "example.com" {
		t.Errorf("TestNewContextHTTP: Expected Host to be lowercased. Given: %s", c.Request.Host)
	}

	if c.Request.URL.Scheme != "https" {
		t.Errorf("TestNewContextHTTP: Expected scheme to be http. Given %s", c.Request.URL.Scheme)
	}

	k.ReturnContext(c)
}
