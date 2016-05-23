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

func TestGetIntSlice(t *testing.T) {
	tests := []struct {
		in     string
		valid  bool
		expect []int
	}{
		{in: "10", valid: true, expect: []int{10}},
		{in: "10,20", valid: true, expect: []int{10, 20}},
		{in: "10, 20", valid: false},
		{in: " 10,20", valid: false},
		{in: "asdfad", valid: false},
		{in: "134254325234", valid: true, expect: []int{134254325234}},
		{in: "2340325,764343,3", valid: true, expect: []int{2340325, 764343, 3}},
		{in: "434,a,3245", valid: false},
		{in: "", valid: false},
	}

	for i, tt := range tests {
		r, _ := http.NewRequest("GET", "/", nil)

		values := url.Values{}
		values.Add("id", tt.in)
		r.URL.RawQuery = values.Encode()

		q := Query{r}

		given, err := q.GetIntSlice("id")
		if tt.valid && err != nil {
			t.Errorf("TestGetIntSlice (%d): Expected valid response. Error: %s", i, err)
		}

		if tt.valid && !reflect.DeepEqual(given, tt.expect) {
			t.Errorf("TestGetIntSlice (%d): Expect %v, given %v", i, tt.expect, given)
		}
	}
}

func TestGetSlice(t *testing.T) {
	tests := []struct {
		in     string
		valid  bool
		expect []string
	}{
		{in: "a", valid: true, expect: []string{"a"}},
		{in: "a,b", valid: true, expect: []string{"a", "b"}},
		{in: "a, b", valid: false},
		{in: " a,b", valid: false},
		{in: "abcdefghijkl", valid: true, expect: []string{"abcdefghijkl"}},
		{in: "2340325,764343,3", valid: true, expect: []string{"2340325", "764343", "3"}},
		{in: "", valid: false},
	}

	for i, tt := range tests {
		r, _ := http.NewRequest("GET", "/", nil)

		values := url.Values{}
		values.Add("names", tt.in)
		r.URL.RawQuery = values.Encode()

		q := Query{r}

		given, err := q.GetSlice("names")
		if tt.valid && err != nil {
			t.Errorf("TestGetSlice (%d): Expected valid response. Error: %s", i, err)
		} else if tt.valid && !reflect.DeepEqual(given, tt.expect) {
			t.Errorf("TestGetSlice (%d): Expect %v, given %v", i, tt.expect, given)
		}
	}
}

func TestSort(t *testing.T) {
	suite := []struct {
		input string
		want  string
	}{
		{input: "a=4&b=&c=10", want: "a=4&c=10"},
		{input: "b=4&a=&d=&c=10&aa=a", want: "aa=a&b=4&c=10"},
		{input: "zed=40&alan=30&sam=32&miChaEl=31", want: "alan=30&miChaEl=31&sam=32&zed=40"},
	}

	for _, s := range suite {
		r, _ := http.NewRequest("GET", "/?"+s.input, nil)

		q := &Query{r}
		given, err := url.QueryUnescape(q.Sort().Encode())
		if err != nil {
			t.Errorf("TestSort: Error unescaping. Err: %s", err)
		}

		if given != s.want {
			t.Errorf("Invalid sort. Expected %q, given %q", s.want, given)
		}
	}
}
