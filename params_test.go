package kumi_test

import (
	"testing"

	"github.com/cristiangraz/kumi"
)

func TestParams(t *testing.T) {
	p := kumi.Params{
		"id":      "10",
		"content": "articles",
		"channel": "tech",
	}

	if id := p.Get("id"); id != "10" {
		t.Fatalf("unexpected id: %s", id)
	} else if content := p.Get("content"); content != "articles" {
		t.Fatalf("unexpected content: %s", content)
	} else if channel := p.Get("channel"); channel != "tech" {
		t.Fatalf("unexpected channel: %s", channel)
	} else if foo := p.Get("foo"); foo != "" {
		t.Fatalf("unexpected foo: %s", foo)
	}

	if id := p.GetDefault("id", "20"); id != "10" {
		t.Fatalf("unexpected id: %s", id)
	} else if fooBar := p.GetDefault("foo", "bar"); fooBar != "bar" {
		t.Fatalf("unexpected fooBar: %s", fooBar)
	}

	if i, err := p.GetInt("id"); err != nil {
		t.Fatalf("error casting to int: %v", err)
	} else if i != 10 {
		t.Fatalf("unexpected id: %d", i)
	}

	if _, err := p.GetInt("channel"); err == nil {
		t.Fatal("expected error casting string to int, none given")
	}
}
