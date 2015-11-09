package middleware

import (
	"io"
	"log"
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

	return func(c *kumi.Context) {
		m := minify.New()
		m.AddFunc("text/css", css.Minify)
		m.AddFunc("text/html", html.Minify)
		m.AddFunc("text/javascript", js.Minify)
		m.AddFunc("application/json", json.Minify)
		m.AddFunc("text/xml", xml.Minify)

		// @todo sync pool
		c.BeforeWrite(func() {
			noTransform := c.Header().Get("Cache-Control")
			if noTransform != "" && strings.Contains(noTransform, "no-transform") {
				return
			}

			contentType := strings.SplitAfterN(c.Header().Get("Content-Type"), ";", 2)[0]
			contentType = strings.Replace(contentType, ";", "", 1)
			if _, ok := allowed[contentType]; !ok {
				return
			}

			pr, pw := io.Pipe()
			go func(w io.Writer) {
				if err := m.Minify(contentType, w, pr); err != nil {
					panic(err)
				}
				log.Println("minified")
			}(c.ResponseWriter)

			c.ResponseWriter = minifyResponseWriter{c.ResponseWriter, pw}
		})

		c.Next()
	}
}
