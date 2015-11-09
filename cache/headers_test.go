package cache

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseCacheControl(t *testing.T) {
	suite := []struct {
		in  string
		out *Headers
	}{
		{in: "", out: NewHeaders()},
		{in: "public", out: NewHeaders().SetPublic()},
		{in: "private", out: NewHeaders().SetPrivate()},
		{in: "must-revalidate", out: NewHeaders().MustRevalidate()},
		{in: "no-cache", out: NewHeaders().NoCache()},
		{in: "public, max-age=30", out: NewHeaders().SetPublic().SetMaxAge(30)},
		{in: "public, max-age=30, s-maxage=10", out: NewHeaders().SetPublic().SetMaxAge(30).SetSharedMaxAge(10)},
		{in: "private, no-cache, no-transform, max-age=30, s-maxage=10", out: NewHeaders().SetPrivate().SetMaxAge(30).SetSharedMaxAge(10).NoCache().NoTransform()},
	}

	for _, s := range suite {
		parsed := parseCacheControl(s.in)
		if !reflect.DeepEqual(s.out, parsed) {
			t.Errorf("TestParse: Expected %s, given %s", s.out, parsed)
		}
	}
}

func TestString(t *testing.T) {
	suite := []struct {
		in  *Headers
		out string
	}{
		{in: NewHeaders(), out: ""},
		{in: NewHeaders().SetPublic(), out: "public"},
		{in: NewHeaders().SetPublic().SetPrivate(), out: "private"},
		{in: NewHeaders().SetMaxAge(30), out: "max-age=30"},
		{in: NewHeaders().SetSharedMaxAge(20), out: "s-maxage=20"},
		{in: NewHeaders().SetMaxAge(30).SetSharedMaxAge(25), out: "max-age=30, s-maxage=25"},
		{in: NewHeaders().NoTransform().NoCache().SetPublic(), out: "no-cache, no-transform, public"},
	}

	for _, s := range suite {
		if s.out != s.in.String() {
			t.Errorf("TestString: Expected %s, given %s", s.out, s.in)
		}
	}
}

func TestEmpty(t *testing.T) {
	h := NewHeaders()
	if !h.IsEmpty() {
		t.Error("TestEmpty: Expected new headers to be empty")
	}

	if h.String() != "" {
		t.Error("TestEmpty: Expected empty headers to return empty string")
	}

	h.NoCache()
	if h.IsEmpty() {
		t.Error("TestEmpty: Expected headers with directive to not be empty")
	}
}

func TestAddRemoveDirectives(t *testing.T) {
	h := NewHeaders()
	if h.Has("public") {
		t.Error("TestAddRemoveDirectives: Expected has public to return false")
	}
	h.AddDirective("public")
	if !h.Has("public") {
		t.Error("TestAddRemoveDirectives: Expected has public to return true")
	}

	h.RemoveDirective("public")
	if !h.IsEmpty() {
		t.Error("TestAddRemoveDirectives: Expected removing only directive would return empty")
	}

	if h.Has("public") {
		t.Error("TestAddRemoveDirectives: Expected has public to return false after removing")
	}
}

func TestSensibleDefault(t *testing.T) {
	suite := []struct {
		headers      map[string]string
		cacheHeaders *Headers
		expected     *Headers
	}{
		{
			cacheHeaders: NewHeaders().SetPublic().SetSharedMaxAge(20),
			expected:     NewHeaders().SetPublic().SetSharedMaxAge(20),
		},
		// If no cache header is defined (Cache-Control, Expires, ETag or Last-Modified),
		// Cache-Control is set to no-cache, meaning that the response will not be cached;
		{
			cacheHeaders: NewHeaders(),
			expected:     NewHeaders().SetPrivate().NoCache(),
		},
		{
			headers:      map[string]string{"Cache-Control": "private"},
			cacheHeaders: NewHeaders(),
			expected:     NewHeaders().SetPrivate(),
		},
		{
			// headers take precedence over cache headers
			headers:      map[string]string{"Cache-Control": "private"},
			cacheHeaders: NewHeaders().SetPublic(),
			expected:     NewHeaders().SetPrivate(),
		},
		{
			// headers take precedence over cache headers
			// Private is still added if public/private/s-maxage not set
			headers:      map[string]string{"Cache-Control": "no-transform"},
			cacheHeaders: NewHeaders().SetPublic().SetMaxAge(30),
			expected:     NewHeaders().NoTransform().SetPrivate(),
		},

		// If Cache-Control is empty (but one of the other cache headers is present), its value is
		// set to private, must-revalidate;
		{
			headers:  map[string]string{"Expires": "Fri, 02 Oct 2015 22:44:20 GMT"},
			expected: NewHeaders().SetPrivate().MustRevalidate(),
		},
		{
			headers:  map[string]string{"ETag": "abc"},
			expected: NewHeaders().SetPrivate().MustRevalidate(),
		},
		{
			headers:  map[string]string{"Last-Modified": "Fri, 02 Oct 2015 22:44:20 GMT"},
			expected: NewHeaders().SetPrivate().MustRevalidate(),
		},

		// But if at least one Cache-Control directive is set, and no public or private directives have
		// been explicitly added, Headers adds the private directive automatically (except when s-maxage is set).
		{
			headers:  map[string]string{"Cache-Control": "no-transform"},
			expected: NewHeaders().NoTransform().SetPrivate(),
		},
		{
			headers:  map[string]string{"Cache-Control": "no-transform, no-cache"},
			expected: NewHeaders().NoTransform().NoCache().SetPrivate(),
		},
		{
			cacheHeaders: NewHeaders().NoTransform(),
			expected:     NewHeaders().NoTransform().SetPrivate(),
		},
		{
			headers:  map[string]string{"Cache-Control": "public, no-transform"},
			expected: NewHeaders().NoTransform().SetPublic(),
		},
		{
			headers:  map[string]string{"Cache-Control": "no-transform, max-age=30"},
			expected: NewHeaders().NoTransform().SetMaxAge(30).SetPrivate(),
		},
		{
			// Private shouldn't be added when s-maxage is set
			headers:      map[string]string{"Cache-Control": "no-transform, s-maxage=30"},
			cacheHeaders: NewHeaders().NoTransform(),
			expected:     NewHeaders().NoTransform().SetSharedMaxAge(30),
		},
		{
			cacheHeaders: NewHeaders().SetPublic(),
			expected:     NewHeaders().SetPublic(),
		},
		{
			cacheHeaders: NewHeaders().SetPrivate(),
			expected:     NewHeaders().SetPrivate(),
		},
		{
			expected: NewHeaders().SetPrivate().NoCache(),
		},
	}

	for i, s := range suite {
		rec := httptest.NewRecorder()
		for name, v := range s.headers {
			rec.Header().Set(name, v)
		}

		if s.cacheHeaders == nil {
			s.cacheHeaders = NewHeaders()
		}

		s.cacheHeaders.SensibleDefaults(rec.Header(), http.StatusOK)
		if rec.Header().Get("Cache-Control") != s.expected.String() {
			t.Errorf("TestSensibleDefault (%d): Expected %s, given %s", i, s.expected, rec.Header().Get("Cache-Control"))
		}
	}
}
