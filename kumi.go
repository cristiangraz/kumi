package kumi

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/justinas/alice"
)

// Engine embeds RouterGroup and provides methods to start the server.
type Engine struct {
	RouterGroup
}

// New creates a new Engine using the given Router.
func New(r Router) *Engine {
	return &Engine{
		RouterGroup: &routerGroup{
			router:     r,
			middleware: alice.New(setup),
		},
	}
}

// Run starts kumi.
func (e *Engine) Run(addr string) error {
	return e.Serve(&ServeConfig{
		Context:          context.Background(),
		InterruptTimeout: 5 * time.Second,
		ContextTimeout:   5 * time.Second,
		Servers: []Server{{
			Server: &http.Server{Addr: addr},
		}},
	})
}

// RunTLS starts kumi with a given TLS config.
func (e *Engine) RunTLS(addr string, config *tls.Config) error {
	return e.Serve(&ServeConfig{
		Context:          context.Background(),
		InterruptTimeout: 5 * time.Second,
		ContextTimeout:   5 * time.Second,
		Servers: []Server{{
			Server: &http.Server{
				Addr:      addr,
				TLSConfig: config,
			},
		}},
	})
}

type ServeConfig struct {
	Context          context.Context
	InterruptTimeout time.Duration
	ContextTimeout   time.Duration
	Servers          []Server
}

type Server struct {
	Server   *http.Server
	Listener net.Listener
}

func (s *Server) serve() error {
	if s.Listener != nil {
		return s.Server.Serve(s.Listener)
	}
	return s.Server.ListenAndServe()
}

// Serve takes one or more http.Server structs and serves those.
// The handler will be set if one is not provided.
func (e *Engine) Serve(config *ServeConfig) error {
	if config == nil {
		return errors.New("config required")
	}
	if config.Context == nil {
		config.Context = context.Background()
	}
	if len(config.Servers) == 0 {
		return errors.New("one or more Servers required")
	}

	// Run servers.
	errch := make(chan error)
	for i := range config.Servers {
		if config.Servers[i].Server.Handler == nil {
			config.Servers[i].Server.Handler = e.RouterGroup
		}
		go func(server Server) {
			if err := server.serve(); err != nil {
				errch <- err
			}
		}(config.Servers[i])
	}

	// Wait for signal.
	graceful := make(chan os.Signal, 1)
	stop := make(chan os.Signal, 1)
	signal.Notify(graceful, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGQUIT)

	var ctx context.Context
	var cancel context.CancelFunc
	select {
	case err := <-errch:
		return err
	case <-graceful: // Signal received. Use parent context.
		ctx, cancel = context.WithTimeout(config.Context, config.InterruptTimeout)
	case <-config.Context.Done(): // Context done. Stop immediately or gracefully shutdown.
		if config.InterruptTimeout == 0 { // Stop immediately.
			for i := range config.Servers {
				config.Servers[i].Server.Close()
			}
			return nil
		}

		// Context was closed, create new context with contextTimeout to
		// set a limit on graceful shutdown.
		ctx, cancel = context.WithTimeout(context.Background(), config.ContextTimeout)
	case <-stop: // Stop immediately.
		for _, server := range config.Servers {
			server.Server.Close()
		}
		return nil
	}
	defer cancel()

	// Graceful shutdown.
	var wg sync.WaitGroup
	for _, server := range config.Servers {
		wg.Add(1)
		go func(server *http.Server) {
			defer wg.Done()
			server.Shutdown(ctx) // Graceful shutdown. Go 1.8 only.
		}(server.Server)
	}

	// Listen for second signal.
	go func() {
		select {
		case <-graceful:
			cancel()
		case <-stop:
			cancel()
		case <-ctx.Done():
		}
	}()

	wg.Wait()

	return nil
}

// setup is internal kumi middleware. It wraps http.ResponseWriter with
// ResponseWriter, or with BodylessResponseWriter for HEAD requests.
// It normalizes the Host and sets the URL scheme. In addition, this
// sets the RequestContext.
func setup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case HEAD:
			w = &BodylessResponseWriter{w}
		default:
			rw := newWriter(w)
			w = rw
			defer writerPool.Put(rw)
		}

		r.Host = strings.ToLower(r.Host)
		if r.TLS != nil {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}

		// Set the kumi request context
		rc := newRequestContext(r)
		defer returnContext(rc)
		if p, ok := getParams(r); ok {
			rc.params = p
		}

		next.ServeHTTP(w, setRequestContext(r, rc))
	})
}
