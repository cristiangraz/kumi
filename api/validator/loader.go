package validator

import (
	"encoding/json"
	"sync"

	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"
)

// JSONLoader is a loader for gojsonschema backed by a sync.Pool.
type JSONLoader struct {
	dst interface{} // The original unmarshaled element
	i   interface{} // needed for validation
}

var ref, _ = gojsonreference.NewJsonReference("#")

// JsonSource implements the interface, but this is not used by gojsonschema.
func (l *JSONLoader) JsonSource() interface{} {
	return nil
}

// LoadJSON marshals and unmarshals the data and returns an interface{} type.
func (l *JSONLoader) LoadJSON() (interface{}, error) {
	if l.i != nil {
		return l.i, nil
	}
	b, err := json.Marshal(l.dst)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &l.i)
	return l.i, err
}

// JsonReference implements the interface.
func (l *JSONLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return ref, nil
}

// LoaderFactory implements the interface, but this is not used by gojsonschema.
func (l *JSONLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return nil
}

var loaderPool = &sync.Pool{
	New: func() interface{} {
		return &JSONLoader{}
	},
}
