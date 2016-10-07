package kumi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristiangraz/kumi"
)

func TestWriter_Status(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			ran = true
			if rw, ok := w.(kumi.ResponseWriter); !ok {
				t.Fatalf("unexpected writer: %T", w)
			} else if rw.Status() != http.StatusPreconditionFailed { // middleware should have access to status
				t.Fatalf("unexpected status code: %d", rw.Status())
			}
		})
	})

	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		rw, _ := w.(kumi.ResponseWriter)

		if rw.Status() != http.StatusOK { // No status should return 200
			t.Fatalf("Expected %d when status not sent, got %d", http.StatusOK, rw.Status())
		}

		w.WriteHeader(http.StatusPreconditionFailed)

		if rw.Status() != http.StatusPreconditionFailed { // Ensure writer is using pointer and status is accessible
			t.Fatalf("Expected %d when status not sent, got %d", http.StatusPreconditionFailed, rw.Status())
		}
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Code != http.StatusPreconditionFailed {
		t.Fatalf("unexpected status code: %d", w.Code)
	}
}

func TestWriter_Written(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("marker"))

		ran = true
		if rw, ok := w.(kumi.ResponseWriter); !ok {
			t.Fatalf("unexpected writer: %T", w)
		} else if rw.Written() != 6 {
			t.Fatalf("unexpected written value: %d", rw.Written())
		}
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", w.Code)
	}
}

// BodylessResponseWriter should not write body or send a Content-Type header.
func TestWriter_NoContentUsesBodylessWriter(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
		w.Header().Set("Content-Type", "application/html")
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("writing content"))
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Code != http.StatusNoContent {
		t.Fatalf("unexpected status code: %d", w.Code)
	} else if w.Body.Len() > 0 {
		t.Fatalf("expected no response body: %s", w.Body.String())
	} else if ct := w.Header().Get("Content-Type"); ct != "" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestWriter_BodylessResponseWriter_Written(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})

	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("writing content"))

		ran = true
		if rw, ok := w.(kumi.ResponseWriter); !ok {
			t.Fatalf("expected kumi.ResponseWriter: %T", w)
		} else if rw.Written() > 0 {
			t.Fatalf("expected no bytes to be written with BodylessResponseWriter, wrote %d", rw.Written())
		}
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

func TestWriter_SetsContentType(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})

	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
		w.Write([]byte("writing content"))
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if ct := w.Header().Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
}

func TestBodyLessResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	bw := &kumi.BodylessResponseWriter{w}

	if n, err := bw.Write([]byte("hi")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if n != 0 {
		t.Fatalf("unexpected number of bytes written: %d", n)
	} else if w.Body.Len() > 0 {
		t.Fatalf("expected no bytes to be written: %s", w.Body.String())
	}
}
