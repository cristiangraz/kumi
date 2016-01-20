package async

import "github.com/cristiangraz/kumi/api"

// Manager holds a dispatcher for managing
// groups of tasks to run in goroutines.
type Manager struct {
	dispatcher *dispatcher
}

// New creates a new Manager with maxWorkers and maxQueue.
func New(maxWorkers, maxQueue int) *Manager {
	m := &Manager{}
	m.dispatcher = newDispatcher(maxWorkers, maxQueue)
	m.dispatcher.run()

	return m
}

// Queue queues up a message to run in the background.
func (m *Manager) Queue(i Invoker, name string, msg *Message) {
	m.Func(name, func() error {
		_, err := i.Invoke(name, msg, true)
		return err
	})
}

// Block runs a blocking function and returns the response.
func (m *Manager) Block(i Invoker, name string, msg *Message) (*api.Response, error) {
	return i.Invoke(name, msg, false)
}

// Func queues up an async function.
func (m *Manager) Func(name string, fn func() error) {
	j := Job{
		Name: name,
		Run:  fn,
	}

	m.dispatcher.jobQueue <- j
}
