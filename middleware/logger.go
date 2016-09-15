package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/cristiangraz/kumi"
)

var logger = &log.Logger{
	Handler: text.New(os.Stderr),
	Level:   log.InfoLevel,
}

// Logger registers the logger.
func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rw, ok := w.(kumi.ResponseWriter)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		defer func() {
			entry := log.NewEntry(logger).WithFields(log.Fields{
				"path":     r.URL.Path,
				"method":   r.Method,
				"status":   rw.Status(),
				"duration": time.Since(start),
			})

			switch {
			case rw.Status() >= 400:
				entry.Warn("")
			default:
				entry.Info("")
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
