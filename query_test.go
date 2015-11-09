package kumi

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestQuery(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("TestQuery: Error creating request. Error: %s", err)
	}

	qs := map[string]string{
		"id":      "25",
		"type":    "articles",
		"channel": "tech",
		"bool1":   "true",
		"bool2":   "false",
		"bool3":   "0",
		"bool4":   "1",
		"bool5":   "2",
	}

	values := url.Values{}
	for name, val := range qs {
		values.Add(name, val)
	}
	r.URL.RawQuery = values.Encode()

	q := Query{r}

	suite := map[string]string{
		"id":      "25",
		"type":    "articles",
		"channel": "tech",
		"foo":     "",
		"bool1":   "true",
		"bool2":   "false",
	}

	for name, expected := range suite {
		if given := q.Get(name); given != expected {
			t.Errorf("TestQuery: Expected Get of %s would equal %s, given %s", name, expected, given)
		}
	}

	if given := q.GetDefault("id", "20"); given != "25" {
		t.Errorf("TestQuery: Expected GetDefault of id would equal 25, given %q", given)
	}

	if given := q.GetDefault("foo", "bar"); given != "bar" {
		t.Errorf("TestQuery: Expected GetDefault of bar would return default value of %q, given %q", "bar", given)
	}

	if given := q.GetDefault("bla", "abc"); given != "abc" {
		t.Errorf("TestQuery: Expected GetDefault of bar would return default value of %q, given %q", "abc", given)
	}

	i, err := q.GetInt("id")
	if err != nil {
		t.Errorf("TestQuery: Expected GetInt of ID would not return an error. Err: %s", err)
	}

	if i != 25 {
		t.Errorf("TestQuery: Expected GetInt of id would return %d, given %d", 25, i)
	}

	_, err = q.GetInt("channel")
	if err == nil {
		t.Error("TestQuery: Expected GetInt of channel would return error. None given.")
	}

	if given := q.All(); !reflect.DeepEqual(r.URL.Query(), given) {
		t.Error("TestQuery: Expected All to return identical url.Values")
	}

	boolSuite := []map[string]string{
		{"name": "bool1", "err": "false", "expected": "true"},
		{"name": "bool2", "err": "false", "expected": "false"},
		{"name": "bool3", "err": "false", "expected": "false"},
		{"name": "bool4", "err": "false", "expected": "true"},
		{"name": "bool5", "err": "true", "expected": "false"},
	}

	for _, s := range boolSuite {
		expected, _ := strconv.ParseBool(s["expected"])
		expectErr, _ := strconv.ParseBool(s["err"])
		given, err := q.GetBool(s["name"])
		if expectErr && err == nil {
			t.Errorf("TestQuery: Expected GetBool would return error. None given.")
		} else {
			if expected != given {
				t.Errorf("TestQuery: Expected GetBool for %s would return %v. Given %v", s["name"], expected, given)
			}
		}
	}
}
