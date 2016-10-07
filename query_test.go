package kumi_test

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi"
)

func TestQuery(t *testing.T) {
	r, _ := http.NewRequest("GET", "/?name=Joe&age=30&foo=true&z=344343&token=OUSFDoshasouBO3325aA", nil)
	q := kumi.NewQuery(r)

	if !reflect.DeepEqual(q.All(), url.Values{
		"name":  {"Joe"},
		"age":   {"30"},
		"foo":   {"true"},
		"z":     {"344343"},
		"token": {"OUSFDoshasouBO3325aA"},
	}) {
		t.Fatalf("unexpected values: %v", q.All())
	} else if q.Get("name") != "Joe" {
		t.Fatalf("unexpected value: %s", q.Get("name"))
	} else if q.GetDefault("bar", "baz") != "baz" {
		t.Fatal("unexpected value")
	} else if age, err := q.GetInt("age"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if age != 30 {
		t.Fatalf("unexpected int value: %d", age)
	} else if b, err := q.GetBool("foo"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if b != true {
		t.Fatalf("Unexpected value: %t", b)
	} else if !reflect.DeepEqual(q.Sort(), url.Values{
		"age":   {"30"},
		"foo":   {"true"},
		"name":  {"Joe"},
		"token": {"OUSFDoshasouBO3325aA"},
		"z":     {"344343"},
	}) {
		t.Fatalf("unexpected value for sort: %v", q.Sort())
	}
}
