package middleware

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/xml"
)

// Minify minifies html, css, and javascript if set as allowed contentTypes.
// If there is an error minifying, it will be skipped and the original bytes
// will be written unaltered.
// If the Cache-Control header is set with a "no-transform" value, nothing
// will be minified.
// Allowed contentTypes: text/css, text/javascript, text/html, text/xml, application/json
func Minify(e *kumi.Engine, contentTypes ...string) {
	allowed := make(map[string]struct{}, len(contentTypes))
	for _, t := range contentTypes {
		allowed[t] = struct{}{}
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("text/xml", xml.Minify)
	m.AddFunc("application/json", json.Minify)

	e.AddListener(kumi.EventFilter, func(c *kumi.Context) {
		noTransform := c.Writer.Header().Get("Cache-Control")
		if noTransform != "" && strings.Contains(noTransform, "no-transform") {
			return
		}

		// Only minify allowed content types
		// If media-type is used (text/plain; charset=utf8), remove it.
		contentType := strings.SplitAfterN(c.Writer.Header().Get("Content-Type"), ";", 2)[0]
		contentType = strings.Replace(contentType, ";", "", 1)
		if _, ok := allowed[contentType]; !ok {
			return
		}

		dst, backup := new(bytes.Buffer), new(bytes.Buffer)
		tee := io.TeeReader(c.Writer, backup)
		if err := m.Minify(contentType, dst, tee); err == nil {
			c.Writer.Replace(dst)
			return
		}

		// On error, finish reading tee until the end
		ioutil.ReadAll(tee)

		c.Writer.Replace(backup)
	})
}
