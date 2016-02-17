package kumi

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type (
	// Query provides useful methods to operate on the request's query string values.
	Query struct {
		request *http.Request
	}
)

var csvIDs = regexp.MustCompile(`^[0-9]+(?:,[0-9]+)*$`)

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
func (q Query) GetInt(name string) (int64, error) {
	return strconv.ParseInt(q.Get(name), 10, 64)
}

// GetIntSlice returns a slice of int64s from a comma-separated list
// of values.
func (q Query) GetIntSlice(name string) ([]int64, error) {
	if q.Get(name) == "" {
		return nil, errors.New("Not found")
	}

	rawIDs := q.Get(name)
	if !csvIDs.MatchString(rawIDs) {
		return nil, errors.New("Invalid csv")
	}

	var slice []int64
	for _, id := range strings.Split(rawIDs, ",") {
		i, _ := strconv.ParseInt(id, 10, 64)
		slice = append(slice, i)
	}

	return slice, nil
}

// GetBool attempts to convert a query string value to a boolean.
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
