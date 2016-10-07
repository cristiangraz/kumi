package kumi

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"sync"
)

// ResponseWriter retains the status code that was written.
type ResponseWriter interface {
	http.ResponseWriter

	// Status returns that status code of the response.
	Status() int

	// Written returns the number of bytes written.
	Written() int
}

type responseWriter struct {
	http.ResponseWriter

	// status holds the status code
	status int

	// wroteHeader tells whether the header's been written.
	wroteHeader bool

	// n holds the number of bytes written.
	n int
}

var _ ResponseWriter = &responseWriter{}

// WriteHeader prepares the response once.If a 204 No Content response
// is being sent, or the BodylessResponseWriter is in use,
// no Content-Type header will be sent.
func (w *responseWriter) WriteHeader(s int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = s

	if s == http.StatusNoContent {
		w.ResponseWriter = &BodylessResponseWriter{w.ResponseWriter}
	}

	if _, ok := w.ResponseWriter.(*BodylessResponseWriter); ok {
		w.Header().Del("Content-Type")
	} else if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/plain")
	}
	w.ResponseWriter.WriteHeader(s)
}

// Writes the response.
func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(p)
	w.n += n
	return n, err
}

// Status returns the status code for the response.
func (w *responseWriter) Status() int {
	return w.status
}

// Written returns the number of bytes written.
func (w *responseWriter) Written() int {
	return w.n
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the response writer doesn't support the http.Hijacker interface")
	}
	return h.Hijack()
}

// Implements the http.CloseNotifier interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// BodylessResponseWriter wraps http.ResponseWriter, discarding
// anything written to the body.
type BodylessResponseWriter struct {
	http.ResponseWriter
}

// Write discards anything written to the body.
func (brw BodylessResponseWriter) Write(b []byte) (int, error) {
	return 0, nil
}

var writerPool = &sync.Pool{
	New: func() interface{} {
		return &responseWriter{}
	},
}

// newWriter returns a new ResponseWriter from the pool.
func newWriter(w http.ResponseWriter) *responseWriter {
	rw := writerPool.Get().(*responseWriter)
	rw.status = http.StatusOK
	rw.ResponseWriter = w
	rw.wroteHeader = false
	rw.n = 0

	return rw
}
