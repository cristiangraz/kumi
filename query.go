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

// Query provides useful methods to operate on the request's query string values.
type Query struct {
	request *http.Request
}

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
func (q Query) GetInt(name string) (int, error) {
	return strconv.Atoi(q.Get(name))
}

// GetIntSlice returns a slice of ints from a comma-separated list
// of values.
func (q Query) GetIntSlice(name string) ([]int, error) {
	raw := q.Get(name)
	if raw == "" {
		return nil, errors.New("not found")
	}

	if !csvIDs.MatchString(raw) {
		return nil, errors.New("invalid csv")
	}

	split := strings.Split(raw, ",")
	slice := make([]int, len(split))
	for i := range split {
		n, _ := strconv.Atoi(split[i])
		slice[i] = n
	}
	return slice, nil
}

// GetSlice returns a slice of strings from a comma-separated list
// of values.
func (q Query) GetSlice(name string) ([]string, error) {
	raw := q.Get(name)
	if raw == "" {
		return nil, errors.New("not found")
	} else if strings.Contains(raw, " ") {
		return nil, errors.New("invalid csv")
	}

	split := strings.Split(raw, ",")
	slice := make([]string, len(split))
	for i := range split {
		slice[i] = split[i]
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
