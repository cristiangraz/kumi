package async

// dispatcher dispatches jobs to workers
type dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	jobQueue   chan Job
}

// newDispatcher creates a dispatcher with a maximum number of workers
func newDispatcher(maxWorkers int, maxQueue int) *dispatcher {
	pool := make(chan chan Job, maxWorkers)

	return &dispatcher{
		workerPool: pool,
		maxWorkers: maxWorkers,
		jobQueue:   make(chan Job, maxQueue),
	}
}

// Run starts the workers
func (d *dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case j := <-d.jobQueue:
			// A job request has been received
			go func(j Job) {
				// Try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.workerPool

				// dispatch the job to the worker job channel
				jobChannel <- j
			}(j)
		}
	}
}
