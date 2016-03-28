package middleware

import (
	"os"
	"time"

	apex "github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/cristiangraz/kumi"
)

var logger = &apex.Logger{
	Handler: text.New(os.Stderr),
	Level:   apex.InfoLevel,
}

// Logger registers the logger.
func Logger(c *kumi.Context) {
	start := time.Now()
	defer func() {
		entry := apex.NewEntry(logger).WithFields(apex.Fields{
			"path":     c.Request.URL.Path,
			"method":   c.Request.Method,
			"status":   c.Status(),
			"duration": time.Since(start),
		})

		if err := kumi.Exception(c); err != nil {
			entry.Errorf("%v", err)
			return
		}

		switch {
		case c.Status() >= 500:
			entry.Error("")
		case c.Status() >= 400:
			entry.Warn("")
		default:
			entry.Info("")
		}
	}()

	c.Next()
}
