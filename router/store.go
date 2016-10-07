package router

// Store ...
type Store struct {
	routes map[string]map[string]bool
}

// Save ...
func (s *Store) Save(method string, pattern string) {
	// Store route
	if s.routes == nil {
		s.routes = make(map[string]map[string]bool, 1)
	}
	if _, ok := s.routes[method]; !ok {
		s.routes[method] = make(map[string]bool, 1)
	}
	s.routes[method][pattern] = true
}

// HasRoute ...
func (s *Store) HasRoute(method string, pattern string) bool {
	if s.routes == nil {
		return false
	} else if routes, ok := s.routes[method]; !ok {
		return false
	} else if _, ok := routes[pattern]; ok {
		return true
	}
	return false
}
