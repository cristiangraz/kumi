package kumi

import (
	"io"

	"golang.org/x/net/context"
)

type (
	// Cacher defines types that can read and write to a cache.
	Cacher interface {
		// Check checks to see if the item is Cacheable and writes the response.
		// If you need data in your Store method, store them in the context.
		Check(*Context) CacheResponse

		// Store stores the compressed content in the reader into cache.
		Store(r io.Reader, c *Context, ttl int) error

		// Purge purges one or more cache keys.
		Purge(keys ...string) error

		// Purge purges all cache keys for a host.
		PurgeAll(host string) error

		// PurgeByTag purges all cache keys for the host with a given tag.
		PurgeByTag(host string, tags ...string) error
	}

	// CacheResponse ...
	CacheResponse interface {
		Found() bool
		Status() int
		Body() io.Reader
		Headers() map[string]string
	}
)

func newContextWithCacheHit(c *Context, hit bool) {
	c.Context = context.WithValue(c.Context, cacheHitKey, hit)
}

func isCacheHit(c *Context) bool {
	h, ok := c.Context.Value(cacheHitKey).(bool)
	if !ok {
		return false
	}

	return h
}

func newContextWithCacheTTL(c *Context, ttl int) {
	c.Context = context.WithValue(c.Context, cacheTTLKey, ttl)
}

func getCacheTTL(c *Context) int {
	i, _ := c.Context.Value(cacheTTLKey).(int)
	return i
}
