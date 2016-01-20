package async

import "log"

type (
	// Job represents the job to be run
	Job struct {
		Name string
		Run  func() error
	}

	// Worker represents the worker that executes the job
	worker struct {
		workerPool chan chan Job
		jobChannel chan Job
		quit       chan bool
	}
)

// A buffered channel that we can send work requests on
var jobQueue chan Job

// NewWorker creates a new worker.
func newWorker(workerPool chan chan Job) worker {
	return worker{
		workerPool: workerPool,
		jobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

// Start method starts the run loop for the worker, listening on a quit channel
// in case we need to stop it
func (w worker) start() {
	go func() {
		for {
			// register the current worker into the worker queue
			w.workerPool <- w.jobChannel

			select {
			case job := <-w.jobChannel:
				log.Printf("Running job with name %q...\n", job.Name)
				if err := job.Run(); err != nil {
					log.Printf("Error running task %q: %s\n", job.Name, err)
				}

			case <-w.quit:
				// We have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests
func (w worker) stop() {
	go func() {
		w.quit <- true
	}()
}
