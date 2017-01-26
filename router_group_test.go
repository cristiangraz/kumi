package kumi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cristiangraz/kumi"
)

func TestRouterGroup_ResponseWriterSet(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
		if _, ok := w.(kumi.ResponseWriter); !ok {
			t.Fatalf("writer is not kumi.ResponseWriter: %T", w)
		} else if _, ok := w.(*kumi.BodylessResponseWriter); ok {
			t.Fatal("unexpected BodylessResponseWriter")
		}
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

func TestRouterGroup_ContextSet(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true

		if name := kumi.Context(r).Query().Get("name"); name != "foo" {
			t.Fatalf("unexpected name: %s", name)
		}
	})

	r, _ := http.NewRequest("GET", "/?name=foo", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

// Test middleware combinations via use, on route, and ordering.
func TestRouterGroup_Middleware_Global(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	var ran bool
	k := kumi.New(&Router{})
	k.Use(a, b, c)
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Body.String() != "abcCBA" {
		t.Fatalf("unexpected order: %s", w.Body.String())
	}
}

// Test middleware combinations via use, on route, and ordering.
func TestRouterGroup_Middleware_GlobalOneByOne(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	var ran bool
	k := kumi.New(&Router{})
	k.Use(c)
	k.Use(b)
	k.Use(a)
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Body.String() != "cbaABC" {
		t.Fatalf("unexpected order: %s", w.Body.String())
	}
}

func TestRouterGroup_Middleware_Local(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	var ran bool
	k := kumi.New(&Router{})
	k.Group(a, b, c).Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Body.String() != "abcCBA" {
		t.Fatalf("unexpected order: %s", w.Body.String())
	}
}

func TestRouterGroup_Middleware_LocalGlobal(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	var ran bool
	k := kumi.New(&Router{})
	k.Use(a, b)
	k.Group(a, b, c).Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Body.String() != "ababcCBABA" {
		t.Fatalf("unexpected order: %s", w.Body.String())
	}
}

func TestRouterGroup_MultipleGroupCalls(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	var ran bool
	k := kumi.New(&Router{})
	k.Use(a, b)
	k.Group(a).Group(b).GroupPath("/users", c).Get("/detail", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("GET", "/users/detail", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	} else if w.Body.String() != "ababcCBABA" {
		t.Fatalf("unexpected order: %s", w.Body.String())
	}
}

func TestRouterGroup_Middleware_LocalGlobalHalt(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")
	halt := tagHaltMiddleware("*")

	var ran bool
	k := kumi.New(&Router{})
	k.Use(a, b)
	k.Group(a, halt, c).Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran {
		t.Fatalf("expected handler not to run")
	} else if w.Body.String() != "aba*ABA" {
		t.Fatalf("unexpected order: %s", w.Body.String())
	}
}

func TestRouterGroup_Methods(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	for _, method := range kumi.HTTPMethods {
		var ran bool
		k := kumi.New(&Router{})
		k.Use(a, c, b)

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ran = true })

		switch method {
		case kumi.GET:
			k.Get("/", h)
		case kumi.HEAD:
			k.Head("/", h)
		case kumi.POST:
			k.Post("/", h)
		case kumi.PUT:
			k.Put("/", h)
		case kumi.PATCH:
			k.Patch("/", h)
		case kumi.OPTIONS:
			k.Options("/", h)
		case kumi.DELETE:
			k.Delete("/", h)
		}

		r, _ := http.NewRequest(method, "/", nil)
		w := httptest.NewRecorder()
		k.ServeHTTP(w, r)

		if !ran {
			t.Fatalf("expected handler to run: %s", method)
		} else if method == kumi.HEAD && w.Body.String() != "" {
			t.Fatalf("no response body should be outputted on HEAD requests: %s", w.Body.String())
		} else if method != kumi.HEAD && w.Body.String() != "acbBCA" {
			t.Fatalf("unexpected middleware order: %s: %s", method, w.Body.String())
		}
	}
}

func TestRouterGroup_All(t *testing.T) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")
	c := tagMiddleware("c")

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(r.Method)) })

	k := kumi.New(&Router{})
	k.Use(a, c, b)
	k.All("/", h)

	for _, method := range kumi.HTTPMethods {
		r, _ := http.NewRequest(method, "/", nil)
		w := httptest.NewRecorder()
		k.ServeHTTP(w, r)

		if method == kumi.HEAD && w.Body.String() != "" {
			t.Fatalf("no response body should be outputted on HEAD requests: %s", w.Body.String())
		} else if method != kumi.HEAD && w.Body.String() != fmt.Sprintf("acb%sBCA", method) {
			t.Fatalf("unexpected middleware order: %s: %s", method, w.Body.String())
		}
	}
}

// Tests that enabling cors automatically creates OPTIONs headers.
func TestRouterGroup_Cors(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.AutoOptionsMethod()
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

// Tests that enabling cors automatically creates OPTIONs headers.
// This verifies that creating a group will maintain the cors setting.
func TestRouterGroup_Cors_Group(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.AutoOptionsMethod()
	k.Group().Get("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

// Tests that enabling cors automatically creates OPTIONs headers.
// This verifies that creating a group will maintain the cors setting.
func TestRouterGroup_Cors_GroupPath(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.AutoOptionsMethod()
	k.GroupPath("/a").Get("/b", func(w http.ResponseWriter, r *http.Request) {
		ran = true
	})

	r, _ := http.NewRequest("OPTIONS", "/a/b", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

func TestRouterGroup_HeadRequestUseBodylessWriter(t *testing.T) {
	var ran bool
	k := kumi.New(&Router{})
	k.Head("/", func(w http.ResponseWriter, r *http.Request) {
		ran = true
		if _, ok := w.(*kumi.BodylessResponseWriter); !ok {
			t.Fatalf("writer is not kumi.BodylessResponseWriter: %T", w)
		}
	})

	r, _ := http.NewRequest("HEAD", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if ran != true {
		t.Fatalf("handler did not run")
	}
}

func TestRouterGroup_HeadAddedToGetRequests(t *testing.T) {
	var ran bool
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ran = true })

	k := kumi.New(&Router{})
	k.Get("/", h)

	r, _ := http.NewRequest("HEAD", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if !ran {
		t.Fatal("expected handler to run")
	}
}

// Not sending an http.Handler should panic.
func TestRouterGroup_NoHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a panic")
		}
	}()

	k := kumi.New(&Router{})
	k.Get("/", nil)

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)
}

// Not sending an http.Handler should panic even with global middleware.
func TestRouterGroup_GlobalMiddlewareNoHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a panic")
		}
	}()

	k := kumi.New(&Router{})
	k.Use(tagMiddleware("*"))
	k.Get("/", nil)

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)
}

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// Each router must implement it's own not found handler. This test is specific to
// our test router, but does verify that the middleware and handler are
// compiled properly.
func TestRouterGroup_NotFoundHandler(t *testing.T) {
	var ran bool
	a := quietMiddleware("a")
	b := quietMiddleware("b")
	c := quietMiddleware("c")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ran = true
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
	})

	k := kumi.New(&Router{})
	k.Use(a, b, c)
	k.NotFoundHandler(h)
	k.Group(tagMiddleware("*")).Get("/some-misc-path", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not run")
	})

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if !ran {
		t.Fatal("not found handler did not run")
	} else if w.Code != http.StatusNotFound {
		t.Fatalf("unexpected code: %d", w.Code)
	} else if w.Body.String() != "404" {
		t.Fatalf("unexpected middleware order: %s", w.Body.String())
	} else if w.Header().Get("a") == "" {
		t.Fatalf("expected a to run")
	} else if w.Header().Get("b") == "" {
		t.Fatalf("expected b to run")
	} else if w.Header().Get("c") == "" {
		t.Fatalf("expected c to run")
	}
}

// Each router must implement it's own method not allowed handler. This test is specific to
// our test router, but does verify that the middleware and handler are
// compiled properly.
func TestRouterGroup_MethodNotAllowedHandler(t *testing.T) {
	var ran bool
	a := quietMiddleware("a")
	b := quietMiddleware("b")
	c := quietMiddleware("c")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ran = true
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method_not_allowed"))
	})

	k := kumi.New(&Router{})
	k.Use(a, b, c)
	k.MethodNotAllowedHandler(h)
	k.Group(tagMiddleware("*")).Get("/", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not run")
	})

	r, _ := http.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if !ran {
		t.Fatal("method not allowed handler did not run")
	} else if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected code: %d", w.Code)
	} else if w.Body.String() != "method_not_allowed" {
		t.Fatalf("unexpected middleware order: %s", w.Body.String())
	} else if w.Header().Get("a") == "" {
		t.Fatalf("expected a to run")
	} else if w.Header().Get("b") == "" {
		t.Fatalf("expected b to run")
	} else if w.Header().Get("c") == "" {
		t.Fatalf("expected c to run")
	}
}

// Router used for testing.
type Router struct {
	routes           map[string]map[string]http.Handler
	notFound         http.Handler
	methodNotAllowed http.Handler
}

func (router *Router) Handle(method string, pattern string, handler http.Handler) {
	if router.routes == nil {
		router.routes = make(map[string]map[string]http.Handler, 1)
	}
	if _, ok := router.routes[method]; !ok {
		router.routes[method] = make(map[string]http.Handler, 1)
	}
	router.routes[method][pattern] = handler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ok := router.HasRoute(r.Method, r.URL.Path); ok {
		h, _ := router.routes[r.Method][r.URL.Path]
		h.ServeHTTP(w, r)
		return
	}

	var methods []string
	for _, m := range kumi.HTTPMethods {
		if router.HasRoute(m, r.URL.Path) {
			methods = append(methods, m)
		}
	}

	if len(methods) > 0 {
		w.Header().Set("Allow", strings.Join(methods, ", "))
		if router.methodNotAllowed != nil {
			router.methodNotAllowed.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if router.notFound != nil {
		router.notFound.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (router *Router) NotFoundHandler(h http.Handler) {
	router.notFound = h
}

func (router *Router) MethodNotAllowedHandler(h http.Handler) {
	router.methodNotAllowed = h
}

func (router *Router) HasRoute(method string, pattern string) bool {
	if router.routes == nil {
		return false
	} else if routes, ok := router.routes[method]; !ok {
		return false
	} else if _, ok := routes[pattern]; ok {
		return true
	}
	return false
}

// A constructor for middleware that writes a "tag" to the ResponseWriter
// for testing middleware ordering. Credit github.com/justinas/alice
// This variation writes the tag before and after to verify middleware flow.
func tagMiddleware(tag string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(strings.ToLower(tag)))
			h.ServeHTTP(w, r)
			w.Write([]byte(strings.ToUpper(tag)))
		})
	}
}

// TagHalt middleware outputs a tag then does not call the next handler
// to halt the middleware chain.
func tagHaltMiddleware(tag string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(strings.ToLower(tag)))
		})
	}
}

// TagHalt middleware sets a header tag to mark it's existence but not send
// a response.
func quietMiddleware(tag string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(tag, "true")
			h.ServeHTTP(w, r)
		})
	}
}
