package cache

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func BenchmarkDefaults(b *testing.B) {
	header := http.Header{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h := New()
		h.SensibleDefaults(header, http.StatusOK)
		Release(h)
	}
}

func TestParseCacheControl(t *testing.T) {
	suite := []struct {
		in  string
		out *Headers
	}{
		{out: New()},
		{in: "public", out: New().SetPublic()},
		{in: "private", out: New().SetPrivate()},
		{in: "must-revalidate", out: New().MustRevalidate()},
		{in: "no-cache", out: New().NoCache()},
		{in: "public, max-age=30", out: New().SetPublic().SetMaxAge(30)},
		{in: "public, max-age=30, s-maxage=10", out: New().SetPublic().SetMaxAge(30).SetSharedMaxAge(10)},
		{in: "private, no-cache, no-transform, max-age=30, s-maxage=10", out: New().SetPrivate().SetMaxAge(30).SetSharedMaxAge(10).NoCache().NoTransform()},
	}

	for _, s := range suite {
		parsed := NewString(s.in)
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
		{in: New(), out: ""},
		{in: New().SetPublic(), out: "public"},
		{in: New().SetPublic().SetPrivate(), out: "private"},
		{in: New().SetMaxAge(30), out: "max-age=30"},
		{in: New().SetSharedMaxAge(20), out: "s-maxage=20"},
		{in: New().SetMaxAge(30).SetSharedMaxAge(25), out: "max-age=30, s-maxage=25"},
		{in: New().NoTransform().NoCache().SetPublic(), out: "no-cache, no-transform, public"},
		{in: New().SetMaxAge(0), out: "max-age=0"},
		{in: New().SetMaxAge(0).SetSharedMaxAge(0), out: "max-age=0, s-maxage=0"},
	}

	for _, s := range suite {
		if s.out != s.in.String() {
			t.Errorf("TestString: Expected %s, given %s", s.out, s.in.String())
		}
	}
}

func TestEmpty(t *testing.T) {
	suite := []struct {
		in    *Headers
		empty bool
	}{
		{in: New(), empty: true},
		{in: New().NoTransform(), empty: false},
		{in: New().NoCache(), empty: false},
		{in: New().MustRevalidate(), empty: false},
		{in: New().SetPublic(), empty: false},
		{in: New().SetPublic().SetPrivate(), empty: false},
		{in: New().SetMaxAge(30), empty: false},
		{in: New().SetSharedMaxAge(20), empty: false},
		{in: New().SetMaxAge(30).SetSharedMaxAge(25), empty: false},
		{in: New().NoTransform().NoCache().SetPublic(), empty: false},
		{in: New().SetMaxAge(0), empty: false},
		{in: New().SetMaxAge(0).SetSharedMaxAge(0), empty: false},
	}

	for i, s := range suite {
		if s.empty != s.in.IsEmpty() {
			t.Errorf("TestEmpty (%d): unexpected: %v", i, s.in.IsEmpty())
		}
		Release(s.in)
	}
}

func TestSensibleDefault(t *testing.T) {
	suite := []struct {
		headers      map[string]string
		cacheHeaders *Headers
		expected     *Headers
	}{
		{
			cacheHeaders: New().SetPublic().SetSharedMaxAge(20),
			expected:     New().SetPublic().SetSharedMaxAge(20),
		},
		// If no cache header is defined (Cache-Control, Expires, ETag or Last-Modified),
		// Cache-Control is set to no-cache, meaning that the response will not be cached;
		{
			cacheHeaders: New(),
			expected:     New().SetPrivate().NoCache(),
		},
		{
			headers:      map[string]string{"Cache-Control": "private"},
			cacheHeaders: New(),
			expected:     New().SetPrivate(),
		},
		{
			// headers take precedence over cache headers
			headers:      map[string]string{"Cache-Control": "private"},
			cacheHeaders: New().SetPublic(),
			expected:     New().SetPrivate(),
		},
		// If Cache-Control is empty (but one of the other cache headers is present), its value is
		// set to private, must-revalidate;
		{
			headers:  map[string]string{"Expires": "Fri, 02 Oct 2015 22:44:20 GMT"},
			expected: New().SetPrivate().MustRevalidate(),
		},
		{
			headers:  map[string]string{"ETag": "abc"},
			expected: New().SetPrivate().MustRevalidate(),
		},
		{
			headers:  map[string]string{"Last-Modified": "Fri, 02 Oct 2015 22:44:20 GMT"},
			expected: New().SetPrivate().MustRevalidate(),
		},

		// But if at least one Cache-Control directive is set, and no public or private directives have
		// been explicitly added, Headers adds the private directive automatically (except when s-maxage is set).
		{
			headers:  map[string]string{"Cache-Control": "no-transform"},
			expected: New().NoTransform().SetPrivate(),
		},
		{
			headers:  map[string]string{"Cache-Control": "no-transform, no-cache"},
			expected: New().NoTransform().NoCache().SetPrivate(),
		},
		{
			cacheHeaders: New().NoTransform(),
			expected:     New().NoTransform().SetPrivate(),
		},
		{
			headers:  map[string]string{"Cache-Control": "public, no-transform"},
			expected: New().NoTransform().SetPublic(),
		},
		{
			headers:  map[string]string{"Cache-Control": "no-transform, max-age=30"},
			expected: New().NoTransform().SetMaxAge(30).SetPrivate(),
		},
		{
			// Private shouldn't be added when s-maxage is set
			headers:      map[string]string{"Cache-Control": "no-transform, s-maxage=30"},
			cacheHeaders: New().NoTransform(),
			expected:     New().NoTransform().SetSharedMaxAge(30),
		},
		{
			cacheHeaders: New().SetPublic(),
			expected:     New().SetPublic(),
		},
		{
			cacheHeaders: New().SetPrivate(),
			expected:     New().SetPrivate(),
		},
		{
			expected: New().SetPrivate().NoCache(),
		},
	}

	for i, s := range suite {
		rec := httptest.NewRecorder()
		for name, v := range s.headers {
			rec.Header().Set(name, v)
		}

		if s.cacheHeaders == nil {
			s.cacheHeaders = New()
		}

		s.cacheHeaders.Parse(rec.Header().Get("Cache-Control"))
		given := s.cacheHeaders.SensibleDefaults(rec.Header(), http.StatusOK)
		if expected := s.expected.String(); given != expected {
			t.Errorf("TestSensibleDefault (%d): Expected %s, given %s", i, expected, given)
		}
		Release(s.cacheHeaders)
		Release(s.expected)
	}
}
