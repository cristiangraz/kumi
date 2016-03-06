# Kumi
Kumi is a very lightweight Go "framework" that packages [x/net/context](https://godoc.org/golang.org/x/net/context),
middleware, and routing. Rather than requiring a specific router, Kumi uses a
router interface so you can choose the router that best suits your project.
Kumi includes three routers by default: httprouter, httptreemux, and gorilla mux.

While Kumi core is light, it does ship with some middleware and functionality
to make developing API endpoints simpler. The API response format is
a subpackage, so you are still free to take a different approach if it makes
more sense for your project.

## Features
 * Fast routing with the flexibility to bring your own router
 * Sub-router and router groups
 * Compatible with ```net/http```
 * Easy access to query params and route params
 * Global middleware and middleware per route group and route
 * Middleware that executes upstream and downstream with the ability to
 stop execution of the next handler.
 * Use of ```x/net/context```. Easy integration with App Engine and default
 context to make unit testing / mocking and environment testing easy.
 * Wrap multiple writers -- including conditional writers based on response headers --
  to avoid buffering. This is how the compression and minification middleware work.
 * Common middleware included. Easily use other Go middleware (including anything
     compatible with ```net/http```).
 * API components (as optional sub-packages) for painless REST API development
 * Built-in CORS handling
 * NotFound and MethodNotAllowed handlers
 * Automatic support for proper HTTP cache headers and included cache sub-package
 for easy reverse proxy/CDN integration.
 * Graceful restarts
 * HTTP/2 support when using TLS
 * Lots of examples and recipes for things like unit testing and API validation/error
 responses.

## Example
```go
func main() {
	router := router.NewHTTPTreeMux()
	k := kumi.New(router)

	// Middleware stack
    k.Use(middleware.Logger)
	k.Use(middleware.Recoverer)
	k.Use(middleware.Compressor)
	k.Use(middleware.Minify)

    k.SetGlobalCors(*kumi.CorsOptions{
        AllowOrigin: []string{"https://dashboard.example.com"},
        MaxAge: time.Duration(24) * time.Hour,
    })

    // Create starting context for all requests
    ctx := context.Background()

	dbConn, err := sql.Open("postgres", "user=foo dbname=bar sslmode=disable")
	if err != nil {
		panic("Connect to db failed")
	}

	// Context: Database
	ctx = WithSQL(ctx, "main", dbConn)
	defer dbConn.Close()

    // Set default Context
    kumi.DefaultContext = ctx

    // auth middleware
    func auth(c *kumi.Context) {
        c.CacheHeaders.SetPrivate()
        if c.Query.Get("key") == "" {
            c.WriteHeader(http.StatusForbidden)
            return
        }

        // Get User and store in context
        var u User
        db := SQL(c.Context, "main")
        if err := db.QueryRow("select id, name from user where key = ?", c.Query.Get("key")).Scan(&u.ID, &u.Name); err != nil {
            c.WriteHeader(http.StatusForbidden)
            return
        }
        c.Context = context.WithValue(c.Context, "user", user)

        c.Next()
    }

	k.Get("/status", func(c *kumi.Context) {
		c.Header().Set("Content-Type", "application/json")
		c.Write([]byte(`{"status": "ok"}`))
	})

    // User area requires auth middleware
    user := k.Group("/user", auth)
    {
        user.Get("/:id", k.Cors("GET", "PUT"), GetUser)
        user.Get("/:id/email", k.Gors("GET"), GetUserEmail)
        user.Put("/:id", k.Cors("GET", "PUT"), UpdateUser)
    }
}
```

## What Kumi is not
Kumi is our own internal starting point for many of our Go packages. We decided to
publish it with the hopes that others would find it useful and hopefully generate
some community contributions in the process. It's not intended to be a full-featured
framework that includes everything everyone would need. If that's what you're looking
for, Kumi likely isn't for you. Kumi is just our take of adding some flexibility
and simple conveniences around net/http by grouping context, middleware, and routing.
That's it.

Take a look at the examples and recipes to get a feel for what Kumi can do.

## Included Routers
While you can easily use your own router by just implementing the Router interface,
the following routers ship with Kumi:

 * github.com/julienschmidt/httprouter
 * github.com/dimfeld/httptreemux
 * github.com/gorilla/mux

If you're unsure which router to use, we recommend starting with httprouter as it
tends to be the fastest and the simplest. If you find you have conflicting routes and
need some more flexibility, httptreemux is a great middle ground. If you need regex routes,
you should go with gorilla.

## Included Middleware
 * Logger: Basic request logging
 * AppEngine: Wraps ```x/net/context``` on each request to get an App Engine compatible
 context.
 * Recoverer: Recovers from panics
 * Security: Create assertion middleware functions to grant or forbid access.
 * Compressor: gzip compression

## Dependencies
 Kumi core has the following dependencies outside of the standard library:

  * golang.org/x/net/context
  * golang.org/x/net/http2 (Will be part of Go core in Go 1.6)
  * github.com/facebookgo/grace/gracehttp

## Contributing
  If you have questions or feel like something is missing that fits in with the goal of Kumi,
  please open up an issue and discuss before submitting a pull request.

  If you want to submit middleware that makes sense to ship in the Kumi middleware
  package, please open up a PR with unit tests.

  If you find a bug or a better way of doing something, please submit a PR! No need to
  open up an issue first. Just be sure to include unit tests!

## Inspiration:
Kumi is a culmination of a lot of great ideas and projects. Here is
a list of great projects you should check out if you like the concepts in Kumi:

 * [kami](https://github.com/guregu/kami): Original inspiration and much of the
 great use of x/net/context. Also partial inspiration for the name (the actual
 name is from a friend's business).
 * [echo](https://github.com/labstack/echo): Sync pools inspiration for the Context
 and compress middleware.
 * [grace](https://github.com/facebookgo/grace): Graceful restarts.
 * [http2](https://godoc.org/golang.org/x/net/http2): Go HTTP/2.
 * [volatile](https://github.com/volatile/core): Very nice framework. Beautiful
 log format and color-coding that was used in the Logger middleware.
 * [alice](https://github.com/justinas/alice): Painless middleware chaining.
