package kumi

import (
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
)

// Query provides useful methods to operate on the request's query string values.
type Query struct {
	request *http.Request
}

var csvIDs = regexp.MustCompile(`^[0-9]+(?:,[0-9]+)*$`)

// NewQuery creates a new Query from a http.Request.
func NewQuery(r *http.Request) *Query {
	return &Query{request: r}
}

// All returns the url.Values from the request's query string.
func (q Query) All() url.Values {
	return q.request.URL.Query()
}

// Get returns a specific query string value.
func (q Query) Get(name string) string {
	return q.request.URL.Query().Get(name)
}

// GetDefault looks for a specific query string value. If that value
// does not exist or is empty, the defaultValue is returned instead.
func (q Query) GetDefault(name string, defaultValue string) string {
	if v := q.Get(name); v != "" {
		return v
	}
	return defaultValue
}

// GetInt attempts to convert a query string value to an integer.
func (q Query) GetInt(name string) (int, error) {
	return strconv.Atoi(q.Get(name))
}

// GetBool returns the boolean value represented by the string.
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
// Any other value returns an error.
func (q Query) GetBool(name string) (bool, error) {
	return strconv.ParseBool(q.Get(name))
}

// Sort returns the query string sorted with empty values removed.
func (q *Query) Sort() url.Values {
	var keys []string
	sorted := url.Values{}
	m := make(map[string]string, len(q.request.URL.Query()))
	for k, v := range q.request.URL.Query() {
		if v[0] == "" {
			continue
		}

		keys = append(keys, k)
		m[k] = v[0]
	}

	sort.Strings(keys)
	for _, k := range keys {
		sorted.Add(k, m[k])
	}
	return sorted
}
