# Kumi
Kumi is a lightweight Go "framework" that packages [x/net/context](https://godoc.org/golang.org/x/net/context),
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
 * API components (as optional sub-packages) for faster API development
 * Built-in CORS handling
 * NotFound and MethodNotAllowed handlers
 * Automatic support for proper HTTP cache headers and included cache sub-package
 for easy reverse proxy/CDN integration.
 * Graceful restarts
 * HTTP/2 support when using TLS

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

## Sub-packages
The following sub-packages are included with Kumi for optional use:

### api
The API package simplifies sending JSON or XML responses, validating incoming JSON request bodies with JSON schema, and converting JSON schema errors to standardized API errors. You can also set an ```io.LimitReader``` to limit the max size of requests. Custom formatters can be created to change the output or output format. This package includes ```formatter``` and ```validator``` sub-packages.

### async
*Experimental*. Allows for queueing/executing asynchronous tasks. There is a channel system that queues jobs and another that controls how many workers operate on the job queue concurrently. This functionality is based on the article: [Handling 1 million requests per minute with golang](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/).

Additionally, there is an ```Invoke``` interface and an AWS Lambda implementation. It allows for invoking Lambda functions synchronously or asynchronously.

### cache
The Cache package makes it easy to set or parse proper ```Cache-Control``` headers. Kumi automatically incorporates this package and will use the ```cache.SensibleDefaults``` function to set defaults (which for the most part means all responses will be set to ```private``` unless explicitly set to something else).

### middleware
Standard middleware functions.

 * Logger: Basic request logging
 * AppEngine: Wraps ```x/net/context``` on each request to get an App Engine compatible
 context.
 * Recoverer: Recovers from panics
 * Security: Create assertion middleware functions to grant or forbid access.
 * Compressor: gzip compression
 * Minify: Minify HTML/CSS/JS/JSON responses

### router
The router package includes router implementations that implement the ```Router``` interface in Kumi. This ensures you can use one of the included routers (see below) or create your own without adjusting your implementation. The benefits are the following items (regardless of if the router specifically implements these features):

 * Router groups
 * Upstream/downstream middleware
 * NotFound and MethodNotAllowed handlers
 * CORS support

#### Included Routers
While you can easily use your own router by just implementing the ```Router``` interface, the following routers ship with Kumi:

 * [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
 * [github.com/dimfeld/httptreemux](https://github.com/dimfeld/httptreemux)
 * [github.com/gorilla/mux](https://github.com/gorilla/mux)

If you're unsure which router to use, we recommend starting with httprouter as it
tends to be the fastest and the simplest. If you find you have conflicting routes and
need some more flexibility, httptreemux is a great middle ground. If you need regex routes,
you should go with gorilla.

## Dependencies
 Kumi core has the following dependencies outside of the standard library:

  * [golang.org/x/net/context](https://golang.org/x/net/context)
  * [github.com/facebookgo/grace/gracehttp](github.com/facebookgo/grace/gracehttp)

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
 * [alice](https://github.com/justinas/alice): Painless middleware chaining.
