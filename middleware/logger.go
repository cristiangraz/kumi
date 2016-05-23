package middleware

import (
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
func Logger(c *kumi.Context) {
	start := time.Now()
	defer func() {
		entry := log.NewEntry(logger).WithFields(log.Fields{
			"path":     c.Request.URL.Path,
			"method":   c.Request.Method,
			"status":   c.Status(),
			"duration": time.Since(start),
		})

		if err := kumi.Exception(c); err != nil {
			entry.Errorf("recovered from panic: %v", err)
			return
		}

		switch {
		case c.Status() >= 400:
			entry.Warn("")
		default:
			entry.Info("")
		}
	}()

	c.Next()
}
