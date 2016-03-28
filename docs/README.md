# Kumi Documentation
Some initial documentation to get started. We'll walk through a sample project step by step. Let's build an API that takes strings and performs string operations. The API will look like this:

## Defining the API
The API will have a single endpoint that changes a string by taking an ```action``` and returning the resulting string.

```POST /change```

```json
{
    "str": "JOHN",
    "action": "lower"
}
```

And will respond with this:
```json
{
    "str": "john"
}
```

Valid actions: ```upper``` or ```lower```

## main.go
Create a main.go file:

```go
package main

import (
    "github.com/cristiangraz/kumi"
    "github.com/cristiangraz/kumi/api"
    "github.com/cristiangraz/kumi/middleware"
    "github.com/cristiangraz/kumi/router"
)

// HTTP method constants for CORS
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	PATCH  = "PATCH"
	DELETE = "DELETE"
)

func main() {

}
```

## Requirements
The API should suppor the following:
 * JSON requests and responses -- including for not found and method not allowed responses
 * Request validation and detailed error messages
 * Request logging to the terminal
 * Recover from panics
 * gzip compression
 * minification of JSON responses only
 * CORS support so this can run across domains in a browser

## Initialize Kumi and define middleware
First step is to select a router and initialize middleware. Because we don't have any route conflicts, we'll stick with httprouter.

```go
k := kumi.New(router.NewHTTPRouter())

k.Use(middleware.Logger)
k.Use(middleware.Recoverer)
k.Use(middleware.Compressor)
k.Use(middleware.MinifyTypes("application/json"))
k.SetGlobalCors(&kumi.CorsOptions{
	AllowOrigin: []string{"*"},
	MaxAge:      time.Duration(24) * time.Hour,
})

// Tell the API we want JSON responses. A formatter is required.
api.Formatter = formatter.JSON
```

So first we create a new instance of Kumi and give it a router. We then configure global middleware.
 * Logger logs requests to the terminal
 * Recoverer recovers from panics
 * Compressor adds gzip compression support
 * MinifyTypes specifies we are going to minify all ```application/json``` responses
 * SetGlobalCors sets CORS options at a global level. Here we are allowing all origins and letting those preflight requests stay cached for 24 hours.

## Error handling
Next lets setup our error handling. We are going to define some basic errors for our API by using the api subpackage. We'll create an ```api.ErrorCollection``` that has all of our errors defined for easy use later.

```go
// API Errors.
const (
 	InvalidJSONError         = "invalid_json"
 	RequestBodyRequiredError = "request_body_required"
 	RequestBodyExceededError = "request_body_exceeded"
 	AccessDeniedError        = "access_denied"
 	NotFoundError            = "not_found"
    MethodNotAllowedError    = "method_not_allowed"
 	InvalidContentTypeError  = "invalid_content_type"
 	RequiredError            = "required"
 	InvalidParameterError    = "invalid_parameter"
 	InvalidParametersError   = "invalid_parameters"
 	UnknownParameterError    = "unknown_parameter"
 	InvalidValueError        = "invalid_value"
 	BadRequestError          = "bad_request"
 	InternalServerError      = "server_error"
 	ServiceUnavailableError  = "service_unavailable"
)

 // errorCollection is assigned to api.Errors in init()
var errorCollection = api.ErrorCollection{
	// API Errors
	InvalidJSONError:         {StatusCode: http.StatusBadRequest, Error: api.Error{Type: InvalidJSONError, Message: "Invalid or malformed JSON"}},
	RequestBodyRequiredError: {StatusCode: http.StatusBadRequest, Error: api.Error{Type: RequestBodyRequiredError, Message: "Request sent with no body"}},
	RequestBodyExceededError: {StatusCode: http.StatusBadRequest, Error: api.Error{Type: RequestBodyExceededError, Message: "Request body exceeded"}},
	AccessDeniedError:        {StatusCode: http.StatusForbidden, Error: api.Error{Type: AccessDeniedError, Message: "Access denied"}},
	NotFoundError:            {StatusCode: http.StatusNotFound, Error: api.Error{Type: NotFoundError, Message: "Not found"}},
	MethodNotAllowedError:    {StatusCode: http.StatusMethodNotAllowed, Error: api.Error{Type: MethodNotAllowedError, Message: "Method not allowed"}},
	InvalidContentTypeError:  {StatusCode: http.StatusUnsupportedMediaType, Error: api.Error{Type: InvalidContentTypeError, Message: "Invalid or missing Content-Type header"}},
	RequiredError:            {StatusCode: 422, Error: api.Error{Type: RequiredError, Message: "Required field missing"}},
	InvalidParameterError:    {StatusCode: 422, Error: api.Error{Type: InvalidParameterError, Message: "Field is invalid. See documentation for more details"}},
	InvalidParametersError:   {StatusCode: 422, Error: api.Error{Type: InvalidParametersError, Message: "One or more parameters is invalid"}},
	UnknownParameterError:    {StatusCode: 422, Error: api.Error{Type: UnknownParameterError, Message: "Unknown parameter sent"}},
	InvalidValueError:        {StatusCode: 422, Error: api.Error{Type: InvalidValueError, Message: "The provided value is invalid"}},
	BadRequestError:          {StatusCode: http.StatusBadRequest, Error: api.Error{Type: BadRequestError, Message: "Bad request."}},
	InternalServerError:      {StatusCode: http.StatusInternalServerError, Error: api.Error{Type: InternalServerError, Message: "Internal server error. The error has been logged and we are working on it"}},
	ServiceUnavailableError:  {StatusCode: http.StatusServiceUnavailable, Error: api.Error{Type: ServiceUnavailableError, Message: "Service unavailable. Please try again shortly"}},
}
```

## Define the routes
Next let's create our API endpoint handlers. We need to support POST requests at ```/change``` that accept JSON requests.

```go
k.Post("/change")
```

Now we have a handler, but there are no attached middleware/handlers (beyond the global middleware). So let's first attach a CORS handler to set CORS responses.

```go
k.Post("/change",
    k.Cors(POST))

// Note: If /change supported GET and PUT requests, you would define it like this:
k.Get("/change",
    k.Cors(GET, POST, PUT))
k.Post("/change",
    k.Cors(GET, POST, PUT))
k.Put("/change",
    k.Cors(GET, POST, PUT))
```

## Create your handlers
Now we need an actual handler to process the request. Let's create one:

```go
func changeHandler(c *kumi.Context) {
    // Handler implementation here
}
```

### Understanding ```kumi.Context```
So let's take a quick detour to understand ```kumi.Context``` -- a critical part of Kumi. So ```kumi.Context``` stores contextual values for requests using ```x/net/context```, and implements the ```http.ResponseWriter``` interface. It looks like this:

```go
type Context struct {
    http.ResponseWriter
    Context      context.Context
    Request      *http.Request
    CacheHeaders *cache.Headers
    Query        Query
    Params       Params
}
```

Let's go through each item one-by-one:

#### http.ResponseWriter
Because this is embedded in ```kumi.Context```, it means you can operate on the ```kumi.Context``` exactly as if it were an ```http.ResponseWriter```. Pretty straight forward... let's move on.

#### Context
Context holds an implementation of ```context.Context```. You should use this to store contextual request values. Usually you start with ```context.Background()```, but Kumi allows you to define your own starting context. This is really useful if you need to store a database connection, api clients, etc across all requests. You can set a starting context, then each request will receive a copy of it where you can safely store request-specific data. You would set your starting context like this:

```go
ctx := context.Background()
// ... add your own values to your starting context

// Set your starting context
kumi.DefaultContext = ctx
```

#### Request
A pointer to ```http.Request``` is stored here. Again this is pretty straightforward, so let's move on.

#### CacheHeaders
Holds a pointer to ```cache.Headers```. We'll utilize this later to see how it works, but you can see more documentation in the cache subpackage.

#### Query
Query simplifies accessing query string parameters. See the godoc for more details.

#### Params
Params provides access to route parameters. I.e. the httprouter route ```/hello/:name``` would have a ```name``` param accessible via ```c.Params.Get("name")```.

## Define your validation rules
We will use the ```api/validation``` sub package to map json schema errors to API errors. So first, the schema:

```go
// Define a schema
var schema := gojsonschema.NewStringLoader(`{
    "type": "object",
    "required": ["str", "action"],
    "additionalProperties": false,
    "properties": {
        "str": {
            "type": "string"
        },
        "action": {
            "type": "string",
            "enum": ["lower", "upper"]
        }
    }
}`)
```

```go
// define your validator options
validatorOpts = &validator.Options{
	RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
	RequestBodyExceeded: errorCollection.Get(RequestBodyExceededError),
	InvalidJSON:         errorCollection.Get(InvalidJSONError),
	BadRequest:          errorCollection.Get(BadRequestError),
	Rules: validator.Rules{
		"*": []validator.Mapping{
			{Type: "required", ErrorType: RequiredError, Message: "Required field missing"},
			{Type: "additional_property_not_allowed", ErrorType: UnknownParameterError, Message: "Unknown parameter sent"},
			{Type: "enum", ErrorType: InvalidValueError, Message: "The provided value is invalid"},
			{Type: "number_one_of", ErrorType: InvalidParametersError, Message: "One or more parameters is invalid"},
			{Type: "number_any_of", ErrorType: InvalidParametersError, Message: "One or more parameters is invalid"},
			{Type: "number_all_of", ErrorType: InvalidParametersError, Message: "One or more parameters is invalid"},
			{Type: "*", ErrorType: InvalidParameterError, Message: "Field is invalid. See documentation for more details"},
		},
	},
	Limit:       int64(1<<19) + 1, // Limit request body at 0.5MB
	ErrorStatus: 422,
	Formatter:   formatter.JSON,
}

// Validators maps handler identifiers (names) to validators. These are
// the validation rules for json schemas used to validate API requests.
Validators = validators{
	"change":             validator.NewValidator(schema, validatorOpts, 0),
}
```

A lot happened here. Let's go through it step by step.

First we created a JSON schema. The schema requires an object with ```str``` and ```action``` properties to both be present with no additional properties sent. ```str``` must be a string and ```action``` must be a string that only allows ```lower``` or ```upper``` for the values.

Next we created our ```validator.Options```. The first four lines specify what errors to return if the RequestBodyRequired error occurs, or if the request includes InvalidJSON.

The ```Rules``` map JSON schema rule types to rule responses.

The ```Limit``` value defines a max limit for use with an ```io.LimitedReader```.

```ErrorStatus``` indicates the status code to respond with for errors.

```Formatter``` defines the formatter to use to format error responses.

Finally, ```Validators``` defines the validators by assigning a name to a configuration that can be used to validate a request. In this case, we want to validate our json schema against the validator options.

## Implement the handler
So now that we have all the pieces, let's implement our handler:

```go
func changeHandler(c *kumi.Context) {
    var data struct {
        Name string `json:"name"`
        Action string `json:"action"`
    }

    // Validate the "change" validator and populate the data struct.
    // If there is an error, return. The validator handles sending error responses.
    if !Validators.Valid("change", &data, c) {
		return
	}

    // No errors. Continue.

    var str string
    switch data.Action {
    case "lower":
        str = strings.ToLower(data.Name)
    case "upper":
        str = strings.ToUpper(data.Name)
    }

    // Send the response
    api.Success(map[string]string{"str": str}).Send(c)
}
```

Now just add the handler to your route definition:

```go
k.Post("/change",
    k.Cors(POST), changeHandler)
```

## Start Kumi
```go
log.Fatal(k.Run(":3000"))
```
