package middleware

import (
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/xml"
)

type (
	minifyResponseWriter struct {
		http.ResponseWriter
		io.Writer
	}
)

func (m minifyResponseWriter) Write(b []byte) (int, error) {
	return m.Writer.Write(b)
}

// Minify returns minify middleware that will minify css, html, javascript, and json
var Minify = MinifyTypes("text/css", "text/html", "text/javascript", "application/json", "text/xml")

// MinifyTypes returns a custom minifier.
func MinifyTypes(contentTypes ...string) kumi.HandlerFunc {
	allowed := make(map[string]struct{}, len(contentTypes))
	for _, t := range contentTypes {
		allowed[t] = struct{}{}
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/json", json.Minify)
	m.AddFunc("text/xml", xml.Minify)

	// @todo sync pool
	return func(c *kumi.Context) {
		c.BeforeWrite(func() {
			noTransform := c.Header().Get("Cache-Control")
			if noTransform != "" && strings.Contains(noTransform, "no-transform") {
				return
			}

			if c.Header().Get("Content-Type") == "" {
				return
			}

			ct, _, err := mime.ParseMediaType(c.Header().Get("Content-Type"))
			if err != nil {
				return
			}

			if _, ok := allowed[ct]; !ok {
				return
			}

			pr, pw := io.Pipe()
			go func(w io.Writer) {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("Error minifying file: %q: Err: %s", c.Request.URL.String(), err)
					}
				}()
				defer pr.Close()
				if err := m.Minify(ct, w, pr); err != nil {
					panic(err)
				}
			}(c.ResponseWriter)

			c.ResponseWriter = minifyResponseWriter{c.ResponseWriter, pw}
		})

		c.Next()
	}
}

// MinifyExtension creates a minifier that only minifies matching extensions
// func MinifyExtension(extensions ...string) kumi.HandlerFunc {}
