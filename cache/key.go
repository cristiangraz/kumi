package cache

import (
	"fmt"
	"net/url"
	"strings"
)

var (
	// KeyFormat is the cache key format
	KeyFormat = "{scheme}://{host}{path}{query}"
)

// Key builds a unique cache key based on a URL.
// Query strings are sorted so that different order strings don't cause
// cache misses.
func Key(u *url.URL) string {
	qs := ""
	if len(u.Query()) > 0 {
		m := make(map[string]string)
		for k, v := range u.Query() {
			m[k] = v[0]
		}

		keys, values := sortMap(m, true)
		for i, k := range keys {
			qs += fmt.Sprintf("%s=%s&", k, values[i])
		}

		if qs != "" {
			qs = "?" + strings.TrimSuffix(qs, "&")
		}
	}

	key := strings.Replace(KeyFormat, "{scheme}", u.Scheme, 1)
	key = strings.Replace(key, "{host}", u.Host, 1)
	key = strings.Replace(key, "{path}", u.Path, 1)
	key = strings.Replace(key, "{query}", qs, 1)

	return key
}
