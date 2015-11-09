package kumi

import "testing"

func TestParams(t *testing.T) {
	p := Params{
		"id":      "10",
		"type":    "articles",
		"channel": "tech",
	}

	suite := map[string]string{
		"id":      "10",
		"type":    "articles",
		"channel": "tech",
		"foo":     "",
	}

	for name, expected := range suite {
		if given := p.Get(name); given != expected {
			t.Errorf("TestParams: Expected Get of %s would equal %s, given %s", name, expected, given)
		}
	}

	if given := p.GetDefault("id", "20"); given != "10" {
		t.Errorf(`TestParams: Expected GetDefault of id would equal 10, given %s`, given)
	}

	if given := p.GetDefault("foo", "bar"); given != "bar" {
		t.Errorf(`TestParams: Expected GetDefault of bar would return default value of "bar", given %s`, given)
	}

	if given := p.GetDefault("bla", "abc"); given != "abc" {
		t.Errorf(`TestParams: Expected GetDefault of bar would return default value of "abc", given %s`, given)
	}

	i, err := p.GetInt("id")
	if err != nil {
		t.Errorf("TestParams: Expected GetInt of ID would not return an error. Err: %s", err)
	}

	if i != 10 {
		t.Errorf("TestParams: Expected GetInt of id would return %d, given %d", 10, i)
	}

	_, err = p.GetInt("channel")
	if err == nil {
		t.Error("TestParams: Expected GetInt of channel would return error. None given.")
	}
}
