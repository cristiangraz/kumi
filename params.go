package kumi

import "strconv"

// Params holds router params.
type Params map[string]string

// Get returns a router parameter by name.
func (p Params) Get(name string) string {
	return p[name]
}

// GetDefault looks for a specific router parameter. If that parameter does not
// exist or is empty, defaultValue is returned instead.
func (p Params) GetDefault(name string, defaultValue string) string {
	if v := p.Get(name); v != "" {
		return v
	}

	return defaultValue
}

// GetInt attempts to convert a router param to an integer.
func (p Params) GetInt(name string) (int, error) {
	return strconv.Atoi(p.Get(name))
}
