## Cache
The Cache package makes it easy to set or parse proper ```Cache-Control``` headers.

```go
cache := cache.NewHeaders()
cache.SetPublic() // Cache-Control: public
cache.SetPrivate() // Cache-Control: private

cache.IsPublic() // false
cache.IsPrivate() // true

cache.NoCache().SetMaxAge(3600) // Cache-Control: private; no-cache; max-age: 3600;
cache.SetSharedMaxAge(900) // Cache-Control: private; no-cache; max-age: 3600; s-maxage: 900;

// Add the Cache-Control header
cache.Add(w.Header())

// or ...
w.Header().Set("Cache-Control", cache.String())
```

You can also use the ```SensibleDefaults``` method by passing the ```http.Header``` and the status code.

If ```http.Header``` has ```Cache-Control``` headers set, those will take precedence over anything in Headers. Follow's [Symfony's guidelines](http://symfony.com/doc/current/book/http_cache.html#caching-rules-and-defaults) for defaults. In general though, all responses will be marked as ```private``` unless explicitly set to ```public```.
