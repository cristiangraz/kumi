# Kumi
Kumi is a lightweight net/http wrapper that packages [context](https://golang.org/pkg/context/),
[middleware](https://github.com/justinas/alice), and interchangeable routers routing. Rather than requiring a specific router, Kumi uses a
router interface so you can choose the router that best suits your project.
Kumi includes three routers by default: httprouter, httptreemux, and gorilla mux.

While Kumi core is light, it does ship with some middleware and functionality
to make developing API endpoints simpler. The API response format is
a subpackage, so you are still free to take a different approach if it makes
more sense for your project.

## Requirements
Kumi requires Go 1.8+.

## Features
 * Fast routing with the flexibility to bring your own router
 * Sub-router and router groups
 * Compatible with ```net/http```
 * Easy access to query params and route params
 * Global middleware and middleware per route group and route
 * Middleware that executes upstream and downstream with the ability to
 stop execution of the next handler
 * `net/http` router group with no `interface{}` types
 * API components (as optional sub-packages) for faster API development: Error handling, success responses, and validation
 * Built-in CORS handling
 * NotFound and MethodNotAllowed handlers
 * Graceful restarts (wraps Go 1.8 [`server.Shutdown()`](https://golang.org/pkg/net/http/#Server.Shutdown) with `os.Signal` handling

## API Validation With JSON schema
Examples TBD.


### Middleware
Standard middleware functions.

 * Logger: Basic request logging
 * Recoverer: Recovers from panics
 * Compressor: gzip compression
 * Minify: Minify HTML/CSS/JS/JSON responses

### Router
The router package includes router implementations that implement the ```RouterGroup``` interface in Kumi. This ensures you can use one of the included routers (see below) or create your own without adjusting your implementation. The benefits are the following items (regardless of if the router specifically implements these features):

 * Router groups
 * Upstream/downstream middleware
 * NotFound and MethodNotAllowed handlers
 * CORS support

#### Included Routers
While you can easily use your own router by just implementing the ```RouterGroup``` interface, the following routers ship with Kumi:

 * [github.com/julienschmidt/httprouter](https://github.com/julienschmidt/httprouter)
 * [github.com/dimfeld/httptreemux](https://github.com/dimfeld/httptreemux)
 * [github.com/gorilla/mux](https://github.com/gorilla/mux)
