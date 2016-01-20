package async

import "github.com/cristiangraz/kumi/api"

// Invoker is used to invoke async methods.
type Invoker interface {
	Invoke(name string, msg *Message, async bool) (*api.Response, error)
}
